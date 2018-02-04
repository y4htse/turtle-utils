package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertTurtlePrice(t *testing.T) {
	assert.NotNil(t, "")
	trtl := 2
	currentPrice := CurrentPrice{}
	currentPrice.btcPrice = 0.01
	currentPrice.usdPrice = 0.000000001
	newPrice, err := ConvertTurtlePrice(currentPrice, int64(trtl))
	assert.Nil(t, err)
	assert.NotNil(t, newPrice)
	assert.Equal(t, newPrice.btcPrice, 0.02)
	assert.Equal(t, newPrice.usdPrice, 0.000000002)
}
