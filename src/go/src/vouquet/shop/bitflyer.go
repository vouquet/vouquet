package shop

import (
	"fmt"
	"time"
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

func openBitflyer(conf *BitflyerConf, ctx context.Context) (*BitflyerHandler, error) {
	var key string
	var secret string
	targets := []string{}
	if conf != nil {
		key = conf.ApiKey
		secret = conf.SecretKey
		targets = conf.Targets
	}

	original_targets := []string{}
	for _, t := range targets {
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
	shop *bitflyer.Bitflyer

	targets []string
}

func (self *BitflyerHandler) GetRate() (map[string]Rate, error) {
	rates, err := self.shop.GetRates(self.targets)
	if err != nil {
		return nil, err
	}

	i_rates := make(map[string]Rate)
	for key, val := range rates {
		i_rates[key] = &BitflyerRate{original:val}
	}
	return i_rates, nil
}

func (self *BitflyerHandler) GetPositions(symbol string) ([]Position, error) {
	return nil, fmt.Errorf("cannot use yet")
}

func (self *BitflyerHandler) GetFixes(symbol string) ([]Fix, error) {
	return nil, fmt.Errorf("cannot use yet")
}

func (self *BitflyerHandler) OrderStreamIn(o_type string, symbol string, size float64) error {
	return fmt.Errorf("cannot use yet")
}

func (self *BitflyerHandler) OrderStreamOut(pos Position) error {
	return fmt.Errorf("cannot use yet")
}

func (self *BitflyerHandler) Release() error {
	return self.shop.Close()
}

type BitflyerRate struct {
	original *bitflyer.Rate
}

func (self *BitflyerRate) Ask() float64 {
	return self.original.Ask()
}

func (self *BitflyerRate) Bid() float64 {
	return self.original.Bid()
}

func (self *BitflyerRate) Last() float64 {
	return self.original.Last()
}

func (self *BitflyerRate) Symbol() string {
	return self.original.ProductCode()
}

func (self *BitflyerRate) Time() time.Time {
	return self.original.Time()
}

func (self *BitflyerRate) Volume() float64 {
	return self.original.Volume()
}

func (self *BitflyerRate) High() float64 {
	return float64(0)
}

func (self *BitflyerRate) Low() float64 {
	return float64(0)
}
