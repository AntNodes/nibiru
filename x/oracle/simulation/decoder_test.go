package simulation_test

import (
	"fmt"
	"testing"

	"github.com/NibiruChain/nibiru/x/common"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/NibiruChain/nibiru/x/oracle/keeper"
	sim "github.com/NibiruChain/nibiru/x/oracle/simulation"
	"github.com/NibiruChain/nibiru/x/oracle/types"
)

var (
	delPk      = ed25519.GenPrivKey().PubKey()
	feederAddr = sdk.AccAddress(delPk.Address())
	valAddr    = sdk.ValAddress(delPk.Address())
)

func TestDecodeDistributionStore(t *testing.T) {
	cdc := keeper.MakeTestCodec(t)
	dec := sim.NewDecodeStore(cdc)

	exchangeRate := sdk.NewDecWithPrec(1234, 1)
	missCounter := uint64(23)

	aggregatePrevote := types.NewAggregateExchangeRatePrevote(types.AggregateVoteHash([]byte("12345")), valAddr, 123)
	aggregateVote := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
		{Pair: common.Pair_NIBI_NUSD.String(), ExchangeRate: sdk.NewDecWithPrec(1234, 1)},
		{Pair: common.Pair_ETH_NUSD.String(), ExchangeRate: sdk.NewDecWithPrec(4321, 1)},
	}, valAddr)

	pair := "btc:usd"

	erBytes, err := exchangeRate.Marshal()
	require.NoError(t, err)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: []byte{0x1, 0x2, 0x3, 0x4, 0x5}, Value: erBytes},
			{Key: []byte{0x2, 0x3, 0x4, 0x5, 0x6}, Value: feederAddr.Bytes()},
			{Key: []byte{0x3, 0x4, 0x5, 0x6, 0x7}, Value: cdc.MustMarshal(&gogotypes.UInt64Value{Value: missCounter})},
			{Key: []byte{0x4, 0x3, 0x5, 0x7, 0x8}, Value: cdc.MustMarshal(&aggregatePrevote)},
			{Key: []byte{0x5, 0x6, 0x7, 0x8, 0x9}, Value: cdc.MustMarshal(&aggregateVote)},
			{Key: append([]byte{0x6}, append([]byte(pair), 0x0)...), Value: []byte{}},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"ExchangeRate", fmt.Sprintf("%v\n%v", exchangeRate, exchangeRate)},
		{"FeederDelegation", fmt.Sprintf("%v\n%v", feederAddr, feederAddr)},
		{"MissCounter", fmt.Sprintf("%v\n%v", missCounter, missCounter)},
		{"AggregatePrevote", fmt.Sprintf("%v\n%v", aggregatePrevote, aggregatePrevote)},
		{"AggregateVote", fmt.Sprintf("%v\n%v", aggregateVote, aggregateVote)},
		{"Pairs", fmt.Sprintf("%s\n%s", pair, pair)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
