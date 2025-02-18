package dex

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/NibiruChain/nibiru/x/dex/keeper"
	"github.com/NibiruChain/nibiru/x/dex/types"
)

// InitGenesis initializes the dex module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetNextPoolNumber(ctx, uint64(genState.Params.StartingPoolNumber))
}

// ExportGenesis returns the dex module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
