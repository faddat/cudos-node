package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CudoVentures/cudos-node/simapp"
	"github.com/CudoVentures/cudos-node/testutil/nullify"
	"github.com/CudoVentures/cudos-node/testutil/sample"
	"github.com/CudoVentures/cudos-node/x/addressbook/client/cli"
	"github.com/CudoVentures/cudos-node/x/addressbook/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

type QueryAddressIntegrationTestSuite struct {
	suite.Suite
	network     *network.Network
	addressList []types.Address
}

func TestQueryAddressIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(QueryAddressIntegrationTestSuite))
}

func (s *QueryAddressIntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up query address integration test suite")

	cfg := simapp.NewConfig()
	cfg.NumValidators = 1

	state := types.GenesisState{}
	require.NoError(s.T(), cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < 5; i++ {
		address := types.Address{
			Creator: sample.AccAddress(),
			Network: "BTC",
			Label:   fmt.Sprintf("%d@testdenom", i),
		}
		nullify.Fill(&address)
		state.AddressList = append(state.AddressList, address)
	}
	s.addressList = state.AddressList

	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(s.T(), err)
	cfg.GenesisState[types.ModuleName] = buf

	s.network = network.New(s.T(), cfg)

	_, err = s.network.WaitForHeight(3) // The network is fully initialized after 3 blocks
	s.Require().NoError(err)
}

func (s *QueryAddressIntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down query address integration test suite")
	s.network.Cleanup()
}

func (s *QueryAddressIntegrationTestSuite) TestShowAddress() {
	ctx := s.network.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc    string
		creator string
		network string
		label   string

		args []string
		err  error
		obj  types.Address
	}{
		{
			desc:    "found",
			creator: s.addressList[0].Creator,
			network: s.addressList[0].Network,
			label:   s.addressList[0].Label,

			args: common,
			obj:  s.addressList[0],
		},
		{
			desc:    "not found",
			creator: sample.AccAddress(),
			network: "BTC",
			label:   "0@testdenom",

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		s.T().Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.creator,
				tc.network,
				tc.label,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowAddress(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetAddressResponse
				require.NoError(t, s.network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Address)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.Address),
				)
			}
		})
	}
}

func (s *QueryAddressIntegrationTestSuite) TestListAddress() {
	ctx := s.network.Validators[0].ClientCtx
	request := func(next []byte, offset, limit uint64, total bool) []string {
		args := []string{
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		}
		if next == nil {
			args = append(args, fmt.Sprintf("--%s=%d", flags.FlagOffset, offset))
		} else {
			args = append(args, fmt.Sprintf("--%s=%s", flags.FlagPageKey, next))
		}
		args = append(args, fmt.Sprintf("--%s=%d", flags.FlagLimit, limit))
		if total {
			args = append(args, fmt.Sprintf("--%s", flags.FlagCountTotal))
		}
		return args
	}
	s.T().Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(s.addressList); i += step {
			args := request(nil, uint64(i), uint64(step), false)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListAddress(), args)
			require.NoError(t, err)
			var resp types.QueryAllAddressResponse
			require.NoError(t, s.network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.Address), step)
			require.Subset(t,
				nullify.Fill(s.addressList),
				nullify.Fill(resp.Address),
			)
		}
	})
	s.T().Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(s.addressList); i += step {
			args := request(next, 0, uint64(step), false)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListAddress(), args)
			require.NoError(t, err)
			var resp types.QueryAllAddressResponse
			require.NoError(t, s.network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.Address), step)
			require.Subset(t,
				nullify.Fill(s.addressList),
				nullify.Fill(resp.Address),
			)
			next = resp.Pagination.NextKey
		}
	})
	s.T().Run("Total", func(t *testing.T) {
		args := request(nil, 0, uint64(len(s.addressList)), true)
		out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListAddress(), args)
		require.NoError(t, err)
		var resp types.QueryAllAddressResponse
		require.NoError(t, s.network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
		require.NoError(t, err)
		require.Equal(t, len(s.addressList), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(s.addressList),
			nullify.Fill(resp.Address),
		)
	})
}
