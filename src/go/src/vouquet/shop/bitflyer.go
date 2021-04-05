package shop

import (
	"fmt"
	"context"
)

import (
	"github.com/vouquet/go-bitflyer/bitflyer"
)

var (
	Symbol2Bitflyer map[string]string
)

func init() {
	Symbol2Bitflyer = make(map[string]string)
	Symbol2Bitflyer[BTC2JPY_spt] = bitflyer.PRODUCTCODE_BTC_JPY
	Symbol2Bitflyer[ETH2JPY_spt] = bitflyer.PRODUCTCODE_ETH_JPY
	Symbol2Bitflyer[XRP2JPY_spt] = bitflyer.PRODUCTCODE_XRP_JPY
	Symbol2Bitflyer[XLM2JPY_spt] = bitflyer.PRODUCTCODE_XLM_JPY
	Symbol2Bitflyer[MONA2JPY_spt] = bitflyer.PRODUCTCODE_MONA_JPY
}

func getBitflyerKey(name string) (string, error) {
	key, ok := Symbol2Bitflyer[name]
	if !ok {
		return "", fmt.Errorf("cannot support '%s'", name)
	}

	return key, nil
}

type BitflyerConf struct {
	ApiKey    string
	SecretKey string

	Targets   []string
}

func openBitflyer(conf *BitflyerConf, ctx context.Context) (*BitflyerConf, error) {
	var key string
	var secret string
	targets := []string{}
	if conf != nil {
		key = conf.ApiKey
		secret = conf.SecretKey
		target = conf.Targets
	}

	original_targets := []string{}
	for _, t := range target {
		o_t, err := getBitflyerKey(t)
		if err != nil {
			return nil, err
		}
		original_targets = append(original_targets, o_t)
	}

	shop, err := bitflyer.NewBitflyer(key, secret, ctx)
	if err != nil {
		return nil, err
	}
	return &BitflyerHandler {
		shop: shop,
		targets: original_targets,
	}, nil
}

type BitflyerHandler struct {
	shop *gomocoin.GoMOcoin

	targets []string
}

func (self *BitflerHandler) GetRate() (map[string]Rate, error) {
	rates, err := self.shop.GetRates(targets)
	if err != nil {
		return nil, err
	}

	i_rates := make(map[string]Rate)
	for key, val := range rates {
		i_rates[key] = val
	}
	return i_rates, nil
}

func (self *BitflerHandler) GetPositions(symbol string) ([]Position, error) {
	return nil, fmt.Errorf("cannot use yet")
}

func (self *BitflerHandler) GetFixes(symbol string) ([]Fix, error) {
	return nil, fmt.Errorf("cannot use yet")
}

func (self *BitflerHandler) OrderStreamIn(o_type string, symbol string, size float64) error {
	return fmt.Errorf("cannot use yet")
}

func (self *BitflerHandler) OrderStreamOut(pos Position) error {
	return fmt.Errorf("cannot use yet")
}

func (self *BitflerHandler) Release() error {
	return self.shop.Close()
}
