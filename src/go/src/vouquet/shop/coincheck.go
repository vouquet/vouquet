package shop

import (
	"fmt"
	"context"
)

import (
	"github.com/vouquet/go-coincheck/coincheck"
)

var (
	Symbol2Coincheck map[string]string
)

func coincheckErrorf(s string, msg ...interface{}) error {
	return fmt.Errorf(NAME_COINCHECK + ": "+ s, msg...)
}

func init() {
	Symbol2Coincheck = make(map[string]string)
	Symbol2Coincheck[BTC2JPY_spt] = coincheck.PAIR_BTC_JPY
}

func getCoincheckKey(name string) (string, error) {
	key, ok := Symbol2Coincheck[name]
	if !ok {
		return "", fmt.Errorf("cannot support '%s'", name)
	}

	return key, nil
}

type CoincheckConf struct {
	ApiKey    string
	SecretKey string
}

func openCoincheck(conf *CoincheckConf, ctx context.Context) (*CoincheckHandler, error) {
	var key string
	var secret string
	if conf != nil {
		key = conf.ApiKey
		secret = conf.SecretKey
	}

	shop, err := coincheck.NewCoincheck(key, secret, ctx)
	if err != nil {
		return nil, coincheckErrorf("%s", err)
	}
	return &CoincheckHandler {
		shop: shop,
	}, nil
}

type CoincheckHandler struct {
	shop *coincheck.Coincheck
}

func (self *CoincheckHandler) GetRate() (map[string]Rate, error) {
	rates, err := self.shop.GetRates()
	if err != nil {
		return nil, coincheckErrorf("%s", err)
	}

	i_rates := make(map[string]Rate)
	for key, val := range rates {
		i_rates[key] = val
	}
	return i_rates, nil
}

func (self *CoincheckHandler) GetPositions(symbol string) ([]Position, error) {
	return nil, coincheckErrorf("cannot use yet")
}

func (self *CoincheckHandler) GetFixes(symbol string) ([]Fix, error) {
	return nil, coincheckErrorf("cannot use yet")
}

func (self *CoincheckHandler) OrderStreamIn(o_type string, symbol string, size float64) error {
	return coincheckErrorf("cannot use yet")
}

func (self *CoincheckHandler) OrderStreamOut(pos Position) error {
	return coincheckErrorf("cannot use yet")
}

func (self *CoincheckHandler) Release() error {
	if err := self.shop.Close(); err != nil {
		return coincheckErrorf("%s", err)
	}
	return nil
}
