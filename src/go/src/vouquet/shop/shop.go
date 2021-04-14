package shop

import (
	"fmt"
	"context"
)

const (
	TYPE_SELL string = "SELL"
	TYPE_BUY  string = "BUY"

	NAME_GMOCOIN   string = "coinzcom"
	NAME_BITFLYER  string = "bitflyer"
	NAME_COINCHECK string = "coincheck"
//	NAME_BINANCE   string = "binance"
)

var (
	NAMES []string = []string{
			NAME_GMOCOIN,
			NAME_BITFLYER,
			NAME_COINCHECK,
//			NAME_BINANCE,
		}
)

type Handler interface {
	GetRate()   (map[string]Rate, error)
	GetPositions(string) ([]Position, error)
	GetFixes(string) ([]Fix, error)
//	OrderStreamIn(string, string, float64) error
//	OrderStreamOut(Position) error
	Order(string, string, float64, bool, float64) error //o_type, symbol, size, isStream, price
	OrderFix(Position, bool, float64) error //Position, isStream, price
//	Order(o_type, string, float64) error //o_type, symbol, size, isStream, price
//	OrderFix(Position) error//Position, isStream, price

	Release() error
}

func New(shop_name string, conf Conf, ctx context.Context) (Handler, error) {
	switch shop_name {
	case NAME_GMOCOIN:
		var c *GmoConf
		if conf != nil {
			c = conf.Gmo()
		}
		return openGmo(c, ctx)
	case NAME_BITFLYER:
		var c *BitflyerConf
		if conf != nil {
			c = conf.Bitflyer()
		}
		return openBitflyer(c, ctx)
	case NAME_COINCHECK:
		var c *CoincheckConf
		if conf != nil {
			c = conf.Coincheck()
		}
		return openCoincheck(c, ctx)
	default:
		break
	}
	return nil, fmt.Errorf("undefined name of shop. '%s'", shop_name)
}
