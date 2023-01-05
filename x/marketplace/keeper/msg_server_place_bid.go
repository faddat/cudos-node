package keeper

import (
	"context"
	"strconv"

	"github.com/CudoVentures/cudos-node/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) PlaceBid(goCtx context.Context, msg *types.MsgPlaceBid) (*types.MsgPlaceBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.PlaceBid(
		ctx,
		msg.AuctionId,
		types.Bid{
			Amount: msg.Amount,
			Bidder: msg.Bidder,
		},
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventPlaceBidType,
			sdk.NewAttribute(types.AttributeAuctionID, strconv.FormatUint(msg.AuctionId, 10)),
			sdk.NewAttribute(types.AttributeKeyPrice, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeBidder, msg.Bidder),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Bidder),
		),
	})

	return &types.MsgPlaceBidResponse{}, nil
}
