package shop

import (
	"fmt"
	"context"
)

const (
	TYPE_SELL string = "SELL"
	TYPE_BUY  string = "BUY"
)

var (
	NAMES []string = []string{GMOCOIN}
)

type Handler interface {
	GetRate()   (map[string]Rate, error)
	GetPositions(string) ([]Position, error)
	GetFixes(string) ([]Fix, error)
	OrderStreamIn(string, string, float64) error
	OrderStreamOut(Position) error
//	Order(o_type, *Symbol, float64)
//	OrderFix

	Release() error //Close() error//TODO: release
}

func New(shop_name string, conf Conf, ctx context.Context) (Handler, error) {
	switch shop_name {
	case GMOCOIN:
		var c *GmoConf
		if conf != nil {
			c = conf.Gmo()
		}
		return openGmo(c, ctx)
	default:
		break
	}
	return nil, fmt.Errorf("undefined name of shop. '%s'", shop_name)
}
