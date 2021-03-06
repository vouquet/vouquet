package shop

import (
	"fmt"
	"strings"
)

const (
	BTC2JPY_spt  string = "BTC2JPY_spt"
	BTC2JPY_mgn  string = "BTC2JPY_mgn"
	ETH2JPY_spt  string = "ETH2JPY_spt"
	ETH2JPY_mgn  string = "ETH2JPY_mgn"
	BCH2JPY_spt  string = "BCH2JPY_spt"
	BCH2JPY_mgn  string = "BCH2JPY_mgn"
	LTC2JPY_spt  string = "LTC2JPY_spt"
	LTC2JPY_mgn  string = "LTC2JPY_mgn"
	XRP2JPY_spt  string = "XRP2JPY_spt"
	XRP2JPY_mgn  string = "XRP2JPY_mgn"
	XLM2JPY_spt  string = "XLM2JPY_spt"
	XLM2JPY_mgn  string = "XLM2JPY_mgn"
	MONA2JPY_spt string = "MONA2JPY_spt"
	MONA2JPY_mgn string = "MONA2JPY_mgn"

	SYMBOL_mgn   string = "_mgn"
	SYMBOL_spt   string = "_spt"

	MODE_spot    string = "SPOT"
	MODE_margin  string = "MARGIN"
)

func GetKey(shop_name string, symbol_name string) (string, error) {
	switch shop_name {
	case NAME_GMOCOIN:
		return getGmoKey(symbol_name)
	case NAME_BITFLYER:
		return getBitflyerKey(symbol_name)
	case NAME_COINCHECK:
		return getCoincheckKey(symbol_name)
	default:
		break
	}
	return "", fmt.Errorf("undefined name of shop '%s'", shop_name)
}

func isMargin(symbol_name string) bool {
	return strings.Contains(symbol_name, SYMBOL_mgn)
}
