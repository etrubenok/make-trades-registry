package fetchers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchSymbols(t *testing.T) {
	f := NewBinanceFetcher()
	symbols, err := f.FetchSymbols()
	assert.NoError(t, err)
	assert.NotZero(t, len(symbols.Symbols))

	names := []string{}
	for _, s := range symbols.Symbols {
		names = append(names, s.Symbol)
	}
	assert.Contains(t, names, strings.ToUpper("btcusdt"))
}
