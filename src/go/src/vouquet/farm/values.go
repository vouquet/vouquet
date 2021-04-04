package farm

import (
	"vouquet/shop"
)

var (
	SOIL_ALL = shop.NAMES

	TYPE_BUY = shop.TYPE_BUY
	TYPE_SELL = shop.TYPE_SELL

	DEFAULT_OpeOption *OpeOption
)

func init() {
	DEFAULT_OpeOption = &OpeOption{
		Stream: true,
	}
}
