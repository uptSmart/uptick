package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/UptickNetwork/uptick/x/nft"
)

var _ nft.MsgServer = Keeper{}

// Send implement Send method of the types.MsgServer.
func (k Keeper) Send(goCtx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	owner := k.GetOwner(ctx, msg.ClassId, msg.Id)
	if !owner.Equals(sender) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the owner of nft %s", sender, msg.Id)
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	if err := k.Transfer(ctx, msg.ClassId, msg.Id, receiver); err != nil {
		return nil, err
	}

	_ = ctx.EventManager().EmitTypedEvent(&nft.EventSend{
		ClassId:  msg.ClassId,
		Id:       msg.Id,
		Sender:   msg.Sender,
		Receiver: msg.Receiver,
	})

	return &nft.MsgSendResponse{}, nil
}


func (k Keeper) IssueClass(goCtx context.Context, msg *nft.MsgIssueClass) (*nft.MsgIssueClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &nft.MsgIssueClassResponse{}, k.SaveClass(ctx, nft.Class{
		Id:          msg.Id,
		Name:        msg.Name,
		Symbol:      msg.Symbol,
		Description: msg.Description,
		Uri:         msg.Uri,
		UriHash:     msg.Issuer,
	})
}

func (k Keeper) MintNFT(goCtx context.Context, msg *nft.MsgMintNFT) (*nft.MsgMintNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var (
		receiver sdk.AccAddress
		err      error
	)

	receiver, err = sdk.AccAddressFromBech32(msg.Minter)
	if err != nil {
		return nil, err
	}

	if msg.Receiver != "" {
		receiver, err = sdk.AccAddressFromBech32(msg.Receiver)
		if err != nil {
			return nil, err
		}
	}
	return &nft.MsgMintNFTResponse{}, k.Mint(ctx, nft.NFT{
		ClassId: msg.ClassId,
		Id:      msg.Id,
		Uri:     msg.Uri,
		UriHash: msg.UriHash,
	}, receiver)
}
