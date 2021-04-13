package shop

import "log"

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
	Symbol2BitflyerCurrencyKey map[string]string
)

func bitflyerErrorf(s string, msg ...interface{}) error {
	return fmt.Errorf(NAME_BITFLYER + ": "+ s, msg...)
}

func init() {
	Symbol2Bitflyer = make(map[string]string)
	Symbol2Bitflyer[BTC2JPY_spt] = bitflyer.PRODUCTCODE_BTC_JPY
	Symbol2Bitflyer[ETH2JPY_spt] = bitflyer.PRODUCTCODE_ETH_JPY
	Symbol2Bitflyer[XRP2JPY_spt] = bitflyer.PRODUCTCODE_XRP_JPY
	Symbol2Bitflyer[XLM2JPY_spt] = bitflyer.PRODUCTCODE_XLM_JPY
	Symbol2Bitflyer[MONA2JPY_spt] = bitflyer.PRODUCTCODE_MONA_JPY

	Symbol2BitflyerCurrencyKey = make(map[string]string)
	Symbol2BitflyerCurrencyKey[BTC2JPY_spt] = bitflyer.CURRENCYCODE_BTC
	Symbol2BitflyerCurrencyKey[ETH2JPY_spt] = bitflyer.CURRENCYCODE_ETH
	Symbol2BitflyerCurrencyKey[XRP2JPY_spt] = bitflyer.CURRENCYCODE_XRP
	Symbol2BitflyerCurrencyKey[XLM2JPY_spt] = bitflyer.CURRENCYCODE_XLM
	Symbol2BitflyerCurrencyKey[MONA2JPY_spt] = bitflyer.CURRENCYCODE_MONA
}

func getBitflyerKey(name string) (string, error) {
	key, ok := Symbol2Bitflyer[name]
	if !ok {
		return "", fmt.Errorf("cannot support '%s'", name)
	}

	return key, nil
}

func getBitflyerCurrencyKey(name string) (string, error) {
	key, ok := Symbol2BitflyerCurrencyKey[name]
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
			return nil, bitflyerErrorf("%s", err)
		}
		original_targets = append(original_targets, o_t)
	}

	shop, err := bitflyer.NewBitflyer(key, secret, ctx)
	if err != nil {
		return nil, bitflyerErrorf("%s", err)
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
		return nil, bitflyerErrorf("%s", err)
	}

	i_rates := make(map[string]Rate)
	for key, val := range rates {
		i_rates[key] = &BitflyerRate{original:val}
	}
	return i_rates, nil
}

func (self *BitflyerHandler) GetPositions(symbol string) ([]Position, error) {
	key, err := getBitflyerKey(symbol)
	if err != nil {
		return nil, bitflyerErrorf("%s", err)
	}
	c_key, err := getBitflyerCurrencyKey(symbol)
	if err != nil {
		return nil, bitflyerErrorf("%s", err)
	}

	b, err := self.shop.GetBalance(c_key)
	if err != nil {
		return nil, bitflyerErrorf("%s", err)
	}
	no_fix_val := b.Available

	o_os, err := self.shop.GetOpenOrders(key)
	if err != nil {
		return nil, bitflyerErrorf("%s", err)
	}
	for _, o := range o_os {
		if o.Side != TYPE_SELL {
			continue
		}
		no_fix_val -= float64(o.Price * o.Size)
	}

	pos := []Position{}
	if no_fix_val < float64(0) {
		log.Println("bitflyer.GetPositions: len: ", len(pos))
		return pos, nil
	}

	c_os, err := self.shop.GetClosedOrders(key)
	if err != nil {
		return nil, bitflyerErrorf("%s", err)
	}
	for _, order := range c_os {
		if order.Side != TYPE_BUY {
			continue
		}

		if no_fix_val < float64(order.Price * order.Size) {
			break
		}
		pos = append(pos, &BitflyerPosition{order:order})
		no_fix_val -= float64(order.Price * order.Size)
	}

	if no_fix_val <= float64(0) {
		log.Println("bitflyer.GetPositions: len: ", len(pos))
		return pos, nil
	}

	rates, err := self.shop.GetRates([]string{key})
	if err != nil {
		return nil, bitflyerErrorf("%s", err)
	}
	rate, ok := rates[key]
	if !ok {
		return nil, bitflyerErrorf("cannot get rate.")
	}

	price := rate.Ask()
	size := no_fix_val - (no_fix_val * bitflyer.FEE_TRADE_RATE)
	dummy_order := &bitflyer.Order{
		Id: time.Now().Unix(),
		Product: key,
		Size: size,
		Price: price,
		Side: TYPE_BUY,
	}
	pos = append(pos, &BitflyerPosition{order:dummy_order})
	log.Println("bitflyer.GetPositions: len: ", len(pos))
	log.Printf("bitflyer.GetPositions: size: %f\n", size)
	return pos, nil
}

func (self *BitflyerHandler) GetFixes(symbol string) ([]Fix, error) {
	return nil, bitflyerErrorf("cannot use yet")
	/*
	key, err := getBitflyerKey(symbol)
	if err != nil {
		return nil, err
	}

	os, err := self.shop.GetClosedOrders(key)
	if err != nil {
		return nil, err
	}

	fs := []Fix{}
	for _, o := range os {
		pos = append(pos, &BitflyerPosition{order:o})
	}
	return fs, nil
	*/
}

func (self *BitflyerHandler) OrderStreamIn(o_type string, symbol string, size float64) error {
	if o_type != TYPE_BUY {
		return bitflyerErrorf("cannot support type of order '%s'", o_type)
	}
	key, err := getBitflyerKey(symbol)
	if err != nil {
		return err
	}
	if _, err := self.shop.MarketOrder(key, o_type, size); err != nil {
		return bitflyerErrorf("%s", err)
	}
	return nil
}

func (self *BitflyerHandler) OrderStreamOut(pos Position) error {
	if pos.OrderType() != TYPE_BUY {
		return bitflyerErrorf("cannot support type of order on position.")
	}
	if _, err := self.shop.MarketOrder(pos.Symbol(), TYPE_SELL, pos.Size()); err != nil {
		return bitflyerErrorf("%s", err)
	}
	return nil
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

type BitflyerPosition struct {
	order *bitflyer.Order
}

func (self *BitflyerPosition) Id() string {
	return fmt.Sprintf("%v", self.order.Id)
}

func (self *BitflyerPosition) Symbol() string {
	return self.order.Product
}

func (self *BitflyerPosition) Size() float64 {
	return self.order.Size
}

func (self *BitflyerPosition) Price() float64 {
	return self.order.Price
}

func (self *BitflyerPosition) OrderType() string {
	return self.order.Side
}
