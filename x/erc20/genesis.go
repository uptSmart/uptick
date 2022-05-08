package erc20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/UptickNetwork/uptick/x/erc20/keeper"
	"github.com/UptickNetwork/uptick/x/erc20/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	// ensure erc20 module account is set on genesis
	if acc := accountKeeper.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the erc20 module account has not been set")
	}

	for _, pair := range data.TokenPairs {
		id := pair.GetID()
		k.SetTokenPair(ctx, pair)
		k.SetDenomsMap(ctx, pair.Denoms, id)
		k.SetERC20Map(ctx, pair.GetERC20Contract(), id)
	}
}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:     k.GetParams(ctx),
		TokenPairs: k.GetAllTokenPairs(ctx),
	}
}
