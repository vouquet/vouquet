package shop

import (
	"fmt"
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

	MODE_spot    string = "SPOT"
	MODE_margin  string = "MARGIN"
)

func GetKey(shop_name string, symbol_name string) (string, error) {
	switch shop_name {
	case GMOCOIN:
		return getGmoKey(symbol_name)
	default:
		break
	}
	return "", fmt.Errorf("undefined name of shop '%s'", shop_name)
}
