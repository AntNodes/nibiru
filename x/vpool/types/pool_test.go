package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/nibiru/x/common"
)

func TestPoolHasEnoughQuoteReserve(t *testing.T) {
	pair := common.MustNewAssetPair("BTC:NUSD")

	pool := &VPool{
		Pair:                   pair,
		TradeLimitRatio:        sdk.MustNewDecFromStr("0.9"), // 0.9
		QuoteAssetReserve:      sdk.NewDec(10_000_000),
		BaseAssetReserve:       sdk.NewDec(10_000_000),
		FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.1"),
		MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.1"),
		MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
		MaxLeverage:            sdk.NewDec(15),
	}

	// less that max ratio
	require.True(t, pool.HasEnoughQuoteReserve(sdk.NewDec(8_000_000)))

	// equal to ratio limit
	require.True(t, pool.HasEnoughQuoteReserve(sdk.NewDec(9_000_000)))

	// more than ratio limit
	require.False(t, pool.HasEnoughQuoteReserve(sdk.NewDec(9_000_001)))
}

func TestSetMarginRatioAndLeverage(t *testing.T) {
	pair := common.MustNewAssetPair("BTC:NUSD")

	pool := &VPool{
		Pair:                   pair,
		TradeLimitRatio:        sdk.MustNewDecFromStr("0.9"), // 0.9
		QuoteAssetReserve:      sdk.NewDec(10_000_000),
		BaseAssetReserve:       sdk.NewDec(10_000_000),
		FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.1"),
		MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.1"),
		MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.42"),
		MaxLeverage:            sdk.NewDec(15),
	}

	require.Equal(t, sdk.MustNewDecFromStr("0.42"), pool.MaintenanceMarginRatio)
	require.Equal(t, sdk.MustNewDecFromStr("15"), pool.MaxLeverage)
}

func TestGetBaseAmountByQuoteAmount(t *testing.T) {
	pair := common.MustNewAssetPair("BTC:NUSD")

	tests := []struct {
		name               string
		baseAssetReserve   sdk.Dec
		quoteAssetReserve  sdk.Dec
		quoteAmount        sdk.Dec
		direction          Direction
		expectedBaseAmount sdk.Dec
		expectedErr        error
	}{
		{
			name:               "quote amount zero",
			baseAssetReserve:   sdk.NewDec(1000),
			quoteAssetReserve:  sdk.NewDec(1000),
			quoteAmount:        sdk.ZeroDec(),
			direction:          Direction_ADD_TO_POOL,
			expectedBaseAmount: sdk.ZeroDec(),
		},
		{
			name:               "simple add quote to pool",
			baseAssetReserve:   sdk.NewDec(1000),
			quoteAssetReserve:  sdk.NewDec(1000),
			quoteAmount:        sdk.NewDec(500),
			direction:          Direction_ADD_TO_POOL,
			expectedBaseAmount: sdk.MustNewDecFromStr("333.333333333333333333"),
		},
		{
			name:               "simple remove quote from pool",
			baseAssetReserve:   sdk.NewDec(1000),
			quoteAssetReserve:  sdk.NewDec(1000),
			quoteAmount:        sdk.NewDec(500),
			direction:          Direction_REMOVE_FROM_POOL,
			expectedBaseAmount: sdk.NewDec(1000),
		},
		{
			name:              "too much quote removed results in error",
			baseAssetReserve:  sdk.NewDec(1000),
			quoteAssetReserve: sdk.NewDec(1000),
			quoteAmount:       sdk.NewDec(1000),
			direction:         Direction_REMOVE_FROM_POOL,
			expectedErr:       ErrQuoteReserveAtZero,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pool := &VPool{
				Pair:                   pair,
				TradeLimitRatio:        sdk.MustNewDecFromStr("0.9"), // 0.9
				QuoteAssetReserve:      tc.quoteAssetReserve,
				BaseAssetReserve:       tc.baseAssetReserve,
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.1"),
				MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.1"),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
				MaxLeverage:            sdk.NewDec(15),
			}

			amount, err := pool.GetBaseAmountByQuoteAmount(tc.direction, tc.quoteAmount)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr,
					"expected error: %w, got: %w", tc.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.EqualValuesf(t, tc.expectedBaseAmount, amount,
					"expected quote: %s, got: %s", tc.expectedBaseAmount.String(), amount.String(),
				)
			}
		})
	}
}

func TestGetQuoteAmountByBaseAmount(t *testing.T) {
	pair := common.MustNewAssetPair("BTC:NUSD")

	tests := []struct {
		name                string
		baseAssetReserve    sdk.Dec
		quoteAssetReserve   sdk.Dec
		baseAmount          sdk.Dec
		direction           Direction
		expectedQuoteAmount sdk.Dec
		expectedErr         error
	}{
		{
			name:                "base amount zero",
			baseAssetReserve:    sdk.NewDec(1000),
			quoteAssetReserve:   sdk.NewDec(1000),
			baseAmount:          sdk.ZeroDec(),
			direction:           Direction_ADD_TO_POOL,
			expectedQuoteAmount: sdk.ZeroDec(),
		},
		{
			name:                "simple add base to pool",
			baseAssetReserve:    sdk.NewDec(1000),
			quoteAssetReserve:   sdk.NewDec(1000),
			baseAmount:          sdk.NewDec(500),
			direction:           Direction_ADD_TO_POOL,
			expectedQuoteAmount: sdk.MustNewDecFromStr("333.333333333333333333"),
		},
		{
			name:                "simple remove base from pool",
			baseAssetReserve:    sdk.NewDec(1000),
			quoteAssetReserve:   sdk.NewDec(1000),
			baseAmount:          sdk.NewDec(500),
			direction:           Direction_REMOVE_FROM_POOL,
			expectedQuoteAmount: sdk.NewDec(1000),
		},
		{
			name:              "too much base removed results in error",
			baseAssetReserve:  sdk.NewDec(1000),
			quoteAssetReserve: sdk.NewDec(1000),
			baseAmount:        sdk.NewDec(1000),
			direction:         Direction_REMOVE_FROM_POOL,
			expectedErr:       ErrBaseReserveAtZero,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pool := &VPool{
				Pair:                   pair,
				TradeLimitRatio:        sdk.OneDec(),
				QuoteAssetReserve:      tc.quoteAssetReserve,
				BaseAssetReserve:       tc.baseAssetReserve,
				FluctuationLimitRatio:  sdk.OneDec(),
				MaxOracleSpreadRatio:   sdk.OneDec(),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
				MaxLeverage:            sdk.NewDec(15),
			}

			amount, err := pool.GetQuoteAmountByBaseAmount(tc.direction, tc.baseAmount)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr,
					"expected error: %w, got: %w", tc.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.EqualValuesf(t, tc.expectedQuoteAmount, amount,
					"expected quote: %s, got: %s", tc.expectedQuoteAmount.String(), amount.String(),
				)
			}
		})
	}
}

func TestIncreaseDecreaseReserves(t *testing.T) {
	pair := common.MustNewAssetPair("ATOM:NUSD")

	pool := &VPool{
		Pair:                   pair,
		TradeLimitRatio:        sdk.MustNewDecFromStr("0.9"), // 0.9
		QuoteAssetReserve:      sdk.NewDec(1_000_000),
		BaseAssetReserve:       sdk.NewDec(1_000_000),
		FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.1"),
		MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.1"),
		MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
		MaxLeverage:            sdk.NewDec(15),
	}

	t.Log("decrease quote asset reserve")
	pool.DecreaseQuoteAssetReserve(sdk.NewDec(100))
	require.Equal(t, sdk.NewDec(999_900), pool.QuoteAssetReserve)

	t.Log("increase quote asset reserve")
	pool.IncreaseQuoteAssetReserve(sdk.NewDec(100))
	require.Equal(t, sdk.NewDec(1_000_000), pool.QuoteAssetReserve)

	t.Log("decrease base asset reserve")
	pool.DecreaseBaseAssetReserve(sdk.NewDec(100))
	require.Equal(t, sdk.NewDec(999_900), pool.BaseAssetReserve)

	t.Log("increase base asset reserve")
	pool.IncreaseBaseAssetReserve(sdk.NewDec(100))
	require.Equal(t, sdk.NewDec(1_000_000), pool.BaseAssetReserve)
}

func TestPool_Validate(t *testing.T) {
	type test struct {
		m         *VPool
		expectErr bool
	}

	cases := map[string]test{
		"invalid pair": {
			m: &VPool{
				Pair: common.AssetPair{},
			},
			expectErr: true,
		},

		"invalid trade limit ratio < 0": {
			m: &VPool{
				Pair:            common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio: sdk.NewDec(-1),
			},
			expectErr: true,
		},

		"invalid trade limit ratio > 1": {
			m: &VPool{
				Pair:            common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio: sdk.NewDec(2),
			},
			expectErr: true,
		},

		"quote asset reserve 0": {
			m: &VPool{
				Pair:              common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:   sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve: sdk.ZeroDec(),
			},
			expectErr: true,
		},

		"base asset reserve 0": {
			m: &VPool{
				Pair:              common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:   sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve: sdk.NewDec(1_000_000),
				BaseAssetReserve:  sdk.ZeroDec(),
			},
			expectErr: true,
		},

		"fluctuation < 0": {
			m: &VPool{
				Pair:                  common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:       sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:     sdk.NewDec(1_000_000),
				BaseAssetReserve:      sdk.NewDec(1_000_000),
				FluctuationLimitRatio: sdk.NewDec(-1),
			},
			expectErr: true,
		},

		"fluctuation > 1": {
			m: &VPool{
				Pair:                  common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:       sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:     sdk.NewDec(1_000_000),
				BaseAssetReserve:      sdk.NewDec(1_000_000),
				FluctuationLimitRatio: sdk.NewDec(2),
			},
			expectErr: true,
		},

		"max oracle spread ratio < 0": {
			m: &VPool{
				Pair:                  common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:       sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:     sdk.NewDec(1_000_000),
				BaseAssetReserve:      sdk.NewDec(1_000_000),
				FluctuationLimitRatio: sdk.MustNewDecFromStr("0.10"),
				MaxOracleSpreadRatio:  sdk.NewDec(-1),
			},
			expectErr: true,
		},

		"max oracle spread ratio > 1": {
			m: &VPool{
				Pair:                  common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:       sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:     sdk.NewDec(1_000_000),
				BaseAssetReserve:      sdk.NewDec(1_000_000),
				FluctuationLimitRatio: sdk.MustNewDecFromStr("0.10"),
				MaxOracleSpreadRatio:  sdk.NewDec(2),
			},
			expectErr: true,
		},

		"maintenance ratio < 0": {
			m: &VPool{
				Pair:                   common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:        sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:      sdk.NewDec(1_000_000),
				BaseAssetReserve:       sdk.NewDec(1_000_000),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.10"),
				MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.10"),
				MaintenanceMarginRatio: sdk.NewDec(-1),
			},
			expectErr: true,
		},

		"maintenance ratio > 1": {
			m: &VPool{
				Pair:                   common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:        sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:      sdk.NewDec(1_000_000),
				BaseAssetReserve:       sdk.NewDec(1_000_000),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.10"),
				MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.10"),
				MaintenanceMarginRatio: sdk.NewDec(2),
			},
			expectErr: true,
		},

		"max leverage < 0": {
			m: &VPool{
				Pair:                   common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:        sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:      sdk.NewDec(1_000_000),
				BaseAssetReserve:       sdk.NewDec(1_000_000),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.10"),
				MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.10"),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.10"),
				MaxLeverage:            sdk.MustNewDecFromStr("-0.10"),
			},
			expectErr: true,
		},

		"max leverage too high for maintenance margin ratio": {
			m: &VPool{
				Pair:                   common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:        sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:      sdk.NewDec(1_000_000),
				BaseAssetReserve:       sdk.NewDec(1_000_000),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.10"),
				MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.10"),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.10"), // Equivalent to 10 leverage
				MaxLeverage:            sdk.MustNewDecFromStr("11"),
			},
			expectErr: true,
		},

		"success": {
			m: &VPool{
				Pair:                   common.MustNewAssetPair("btc:usd"),
				TradeLimitRatio:        sdk.MustNewDecFromStr("0.10"),
				QuoteAssetReserve:      sdk.NewDec(1_000_000),
				BaseAssetReserve:       sdk.NewDec(1_000_000),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.10"),
				MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.10"),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
				MaxLeverage:            sdk.MustNewDecFromStr("15"),
			},
			expectErr: false,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := tc.m.Validate()
			if err == nil && tc.expectErr {
				t.Fatal("error expected")
			} else if err != nil && !tc.expectErr {
				t.Fatal("unexpected error")
			}
		})
	}
}

func TestVPool_GetMarkPrice(t *testing.T) {
	tests := []struct {
		name          string
		pool          VPool
		expectedValue sdk.Dec
	}{
		{
			"happy path",
			VPool{
				Pair:              common.Pair_BTC_NUSD,
				BaseAssetReserve:  sdk.MustNewDecFromStr("10"),
				QuoteAssetReserve: sdk.MustNewDecFromStr("10000"),
			},
			sdk.MustNewDecFromStr("1000"),
		},
		{
			"nil base",
			VPool{
				Pair:              common.Pair_BTC_NUSD,
				BaseAssetReserve:  sdk.Dec{},
				QuoteAssetReserve: sdk.MustNewDecFromStr("10000"),
			},
			sdk.ZeroDec(),
		},
		{
			"zero base",
			VPool{
				Pair:              common.Pair_BTC_NUSD,
				BaseAssetReserve:  sdk.ZeroDec(),
				QuoteAssetReserve: sdk.MustNewDecFromStr("10000"),
			},
			sdk.ZeroDec(),
		},
		{
			"nil quote",
			VPool{
				Pair:              common.Pair_BTC_NUSD,
				BaseAssetReserve:  sdk.MustNewDecFromStr("10"),
				QuoteAssetReserve: sdk.Dec{},
			},
			sdk.ZeroDec(),
		},
		{
			"zero quote",
			VPool{
				Pair:              common.Pair_BTC_NUSD,
				BaseAssetReserve:  sdk.MustNewDecFromStr("10"),
				QuoteAssetReserve: sdk.ZeroDec(),
			},
			sdk.ZeroDec(),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, tc.expectedValue.Equal(tc.pool.GetMarkPrice()))
		})
	}
}

func TestVPool_IsOverFluctuationLimit(t *testing.T) {
	tests := []struct {
		name string
		pool VPool

		isOverLimit bool
	}{
		{
			name: "zero fluctuation limit ratio",
			pool: VPool{
				Pair:                   common.Pair_BTC_NUSD,
				QuoteAssetReserve:      sdk.OneDec(),
				BaseAssetReserve:       sdk.OneDec(),
				FluctuationLimitRatio:  sdk.ZeroDec(),
				TradeLimitRatio:        sdk.OneDec(),
				MaxOracleSpreadRatio:   sdk.OneDec(),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
				MaxLeverage:            sdk.MustNewDecFromStr("15"),
			},
			isOverLimit: false,
		},
		{
			name: "lower limit of fluctuation limit",
			pool: VPool{
				Pair:                   common.Pair_BTC_NUSD,
				QuoteAssetReserve:      sdk.NewDec(999),
				BaseAssetReserve:       sdk.OneDec(),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.001"),
				TradeLimitRatio:        sdk.OneDec(),
				MaxOracleSpreadRatio:   sdk.OneDec(),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
				MaxLeverage:            sdk.MustNewDecFromStr("15"),
			},
			isOverLimit: false,
		},
		{
			name: "upper limit of fluctuation limit",
			pool: VPool{
				Pair:                   common.Pair_BTC_NUSD,
				QuoteAssetReserve:      sdk.NewDec(1001),
				BaseAssetReserve:       sdk.OneDec(),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.001"),
				TradeLimitRatio:        sdk.OneDec(),
				MaxOracleSpreadRatio:   sdk.OneDec(),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
				MaxLeverage:            sdk.MustNewDecFromStr("15"),
			},
			isOverLimit: false,
		},
		{
			name: "under fluctuation limit",
			pool: VPool{
				Pair:                   common.Pair_BTC_NUSD,
				QuoteAssetReserve:      sdk.NewDec(998),
				BaseAssetReserve:       sdk.OneDec(),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.001"),
				TradeLimitRatio:        sdk.OneDec(),
				MaxOracleSpreadRatio:   sdk.OneDec(),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
				MaxLeverage:            sdk.MustNewDecFromStr("15"),
			},
			isOverLimit: true,
		},
		{
			name: "over fluctuation limit",
			pool: VPool{
				Pair:                   common.Pair_BTC_NUSD,
				QuoteAssetReserve:      sdk.NewDec(1002),
				BaseAssetReserve:       sdk.OneDec(),
				FluctuationLimitRatio:  sdk.MustNewDecFromStr("0.001"),
				TradeLimitRatio:        sdk.OneDec(),
				MaxOracleSpreadRatio:   sdk.OneDec(),
				MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
				MaxLeverage:            sdk.MustNewDecFromStr("15"),
			},
			isOverLimit: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			snapshot := NewReserveSnapshot(
				common.Pair_BTC_NUSD,
				sdk.OneDec(),
				sdk.NewDec(1000),
				time.Now(),
			)
			assert.EqualValues(t, tc.isOverLimit, tc.pool.IsOverFluctuationLimitInRelationWithSnapshot(snapshot))
		})
	}
}
