package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/UptickNetwork/uptick/x/nft"

	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
)

// Keeper of the nft store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	nftkeeper.Keeper
}

// NewKeeper creates a new instance of the NFT Keeper
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	ak nft.AccountKeeper,
	bk nft.BankKeeper,
) Keeper {

	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}

	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		Keeper:   nftkeeper.NewKeeper(storeKey, cdc, ak, bk),
	}
}
