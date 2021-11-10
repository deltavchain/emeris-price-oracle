package rest_api

import (
	"encoding/json"
	"testing"

	"github.com/allinbits/emeris-price-oracle/price-oracle/types"
	"github.com/stretchr/testify/require"
)

func TestAllPricesHandler(t *testing.T) {
	router, ctx, w, tDown := setup(t)
	defer tDown()

	wantData := types.AllPriceResponse{
		Fiats: []types.FiatPriceResponse{
			{Symbol: "USDCHF", Price: 10},
			{Symbol: "USDEUR", Price: 20},
			{Symbol: "USDKRW", Price: 5},
		},
		Tokens: []types.TokenPriceResponse{
			{Price: 10, Symbol: "ATOMUSDT", Supply: 113563929433.0},
			{Price: 10, Symbol: "LUNAUSDT", Supply: 113563929433.0},
		},
	}
	err := insertWantData(router, wantData, router.s.l)
	require.NoError(t, err)

	router.allPricesHandler(ctx)

	var got struct {
		Data types.AllPriceResponse `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)

	require.Equal(t, wantData, got.Data)
}
