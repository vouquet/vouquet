package farm

import (
	"fmt"
	"context"
)

import (
	"github.com/vouquet/go-gmo-coin/gomocoin"
	"github.com/vouquet/shop"
)

func openShop(name string, conf *Config, ctx context.Context) (shop.Shop, error) {
	var s shop.Shop
	switch name {
	case SOIL_GMO:
		var key string
		var secret string
		if conf != nil {
			key = conf.GMO.ApiKey
			secret = conf.GMO.SecretKey
		}

		g_shop, err := gomocoin.NewGoMOcoin(key, secret, ctx)
		if err != nil {
			return nil, err
		}
		s = g_shop

	default:
		return nil, fmt.Errorf("undefined name of soil '%s'", name)
	}

	return s, nil
}
