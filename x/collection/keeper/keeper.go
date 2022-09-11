package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/UptickNetwork/uptick/x/collection/types"
	"github.com/UptickNetwork/uptick/x/nft"
	nftkeeper "github.com/UptickNetwork/uptick/x/nft/keeper"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	storeKey storetypes.StoreKey // Unexposed key to access store from sdk.Context
	cdc      codec.Codec
	nk       nftkeeper.Keeper
}

// NewKeeper creates a new instance of the NFT Keeper
func NewKeeper(cdc codec.Codec,
	storeKey storetypes.StoreKey,
	ak nft.AccountKeeper,
	bk nft.BankKeeper,
) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		nk:       nftkeeper.NewKeeper(cdc,storeKey,ak, bk),
	}
}

// IssueDenom issues a denom according to the given params
func (k Keeper) IssueDenom(ctx sdk.Context,
	id, name, schema, symbol string,
	creator sdk.AccAddress,
	mintRestricted, updateRestricted bool,
) error {
	denomMetadata := &types.DenomMetadata{
		Creator:          creator.String(),
		Schema:           schema,
		MintRestricted:   mintRestricted,
		UpdateRestricted: updateRestricted,
	}
	data, err := codectypes.NewAnyWithValue(denomMetadata)
	if err != nil {
		return err
	}
	return k.nk.SaveClass(ctx, nft.Class{
		Id:     id,
		Name:   name,
		Symbol: symbol,
		Data:   data,
	})
}

// MintNFT mints an NFT and manages the NFT's existence within Collections and Owners
func (k Keeper) MintNFT(
	ctx sdk.Context,
	denomID string,
	tokenID string,
	tokenNm string,
	tokenURI string,
	tokenData string,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
) error {
	denom, err := k.GetDenomInfo(ctx, denomID)
	if err != nil {
		return err
	}
	if denom.MintRestricted && denom.Creator != sender.String() {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to mint NFT of denom %s", sender, denomID)
	}

	nftMetadata := &types.NFTMetadata{
		Name:        tokenNm,
		Description: tokenData,
	}
	data, err := codectypes.NewAnyWithValue(nftMetadata)
	if err != nil {
		return err
	}

	return k.nk.Mint(
		ctx,
		nft.NFT{
			ClassId: denomID,
			Id:      tokenID,
			Uri:     tokenURI,
			UriHash: "",
			Data:    data,
		},
		receiver,
	)
}

// EditNFT updates an already existing NFT
func (k Keeper) EditNFT(
	ctx sdk.Context, denomID, tokenID, tokenNm,
	tokenURI, tokenData string, owner sdk.AccAddress,
) error {
	denom, err := k.GetDenomInfo(ctx, denomID)
	if err != nil {
		return err
	}

	if denom.UpdateRestricted {
		// if true , nobody can update the NFT under this denom
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "nobody can update the NFT under this denom %s", denomID)
	}

	// just the owner of NFT can edit
	if err := k.Authorize(ctx, denomID, tokenID, owner); err != nil {
		return err
	}

	token, exist := k.nk.GetNFT(ctx, denomID, tokenID)
	if !exist {
		return sdkerrors.Wrapf(types.ErrUnknownNFT, "nft ID %s not exists", tokenID)
	}

	if types.Modified(tokenURI) {
		token.Uri = tokenURI
	}

	if types.Modified(tokenNm) || types.Modified(tokenData) {
		var nftMetadata types.NFTMetadata
		if err := k.cdc.Unmarshal(token.Data.GetValue(), &nftMetadata); err != nil {
			return err
		}

		if types.Modified(tokenNm) {
			nftMetadata.Name = tokenNm
		}

		if types.Modified(tokenData) {
			nftMetadata.Description = tokenData
		}

		data, err := codectypes.NewAnyWithValue(&nftMetadata)
		if err != nil {
			return err
		}
		token.Data = data
	}
	return k.nk.Update(ctx, token)
}

// TransferOwnership transfers the ownership of the given NFT to the new owner
func (k Keeper) TransferOwnership(
	ctx sdk.Context, denomID, tokenID, tokenNm, tokenURI,
	tokenData string, srcOwner, dstOwner sdk.AccAddress,
) error {
	token, exist := k.nk.GetNFT(ctx, denomID, tokenID)
	if !exist {
		return sdkerrors.Wrapf(types.ErrInvalidTokenID, "nft ID %s not exists", tokenID)
	}

	if err := k.Authorize(ctx, denomID, tokenID, srcOwner); err != nil {
		return err
	}

	denom, err := k.GetDenomInfo(ctx, denomID)
	if err != nil {
		return err
	}

	var changed bool
	if types.Modified(tokenURI) {
		token.Uri = tokenURI
		changed = true
	}

	var nftMetadata types.NFTMetadata
	if err := k.cdc.Unmarshal(token.Data.GetValue(), &nftMetadata); err != nil {
		return err
	}

	if types.Modified(tokenNm) {
		nftMetadata.Name = tokenNm
		changed = true
	}

	if types.Modified(tokenData) {
		nftMetadata.Description = tokenData
		changed = true
	}

	if denom.UpdateRestricted && changed {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "It is restricted to update NFT under this denom %s", denom.ID)
	}

	if changed {
		data, err := codectypes.NewAnyWithValue(&nftMetadata)
		if err != nil {
			return err
		}
		token.Data = data
		if err := k.nk.Update(ctx, token); err != nil {
			return err
		}
	}
	return k.nk.Transfer(ctx, denomID, tokenID, dstOwner)
}

// BurnNFT deletes a specified NFT
func (k Keeper) BurnNFT(ctx sdk.Context, denomID, tokenID string, owner sdk.AccAddress) error {
	if err := k.Authorize(ctx, denomID, tokenID, owner); err != nil {
		return err
	}
	return k.nk.Burn(ctx, denomID, tokenID)
}

// TransferDenomOwner transfers the ownership of the given denom to the new owner
func (k Keeper) TransferDenomOwner(
	ctx sdk.Context, denomID string, srcOwner, dstOwner sdk.AccAddress,
) error {
	denom, err := k.GetDenomInfo(ctx, denomID)
	if err != nil {
		return err
	}

	// authorize
	if srcOwner.String() != denom.Creator {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to transfer denom %s", srcOwner.String(), denomID)
	}

	denomMetadata := &types.DenomMetadata{
		Creator:          dstOwner.String(),
		Schema:           denom.Schema,
		MintRestricted:   denom.MintRestricted,
		UpdateRestricted: denom.UpdateRestricted,
	}
	data, err := codectypes.NewAnyWithValue(denomMetadata)
	if err != nil {
		return err
	}
	return k.nk.UpdateClass(ctx, nft.Class{
		Id:     denom.ID,
		Name:   denom.Name,
		Symbol: denom.Symbol,
		Data:   data,
	})
}
