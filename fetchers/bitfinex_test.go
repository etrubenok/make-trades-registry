package fetchers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitfinexFetchSymbols(t *testing.T) {

	f := NewBitfinexFetcher()
	_, err := f.FetchSymbols()
	assert.NoError(t, err)

}
