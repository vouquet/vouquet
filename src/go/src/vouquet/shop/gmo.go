package shop

import (
	"fmt"
	"time"
	"context"
)

import (
	"github.com/vouquet/go-gmo-coin/gomocoin"
)

var (
	Symbol2Gmo       map[string]string
	GmoSymbol2Symbol map[string]string
	Mode2Gmo         map[string]string
)

func gmoErrorf(s string, msg ...interface{}) error {
	return fmt.Errorf(NAME_GMOCOIN + ": "+ s, msg...)
}

func init() {
	Symbol2Gmo = make(map[string]string)
	Symbol2Gmo[BTC2JPY_spt] = gomocoin.SYMBOL_BTC
	Symbol2Gmo[ETH2JPY_spt] = gomocoin.SYMBOL_ETH
	Symbol2Gmo[BCH2JPY_spt] = gomocoin.SYMBOL_BCH
	Symbol2Gmo[LTC2JPY_spt] = gomocoin.SYMBOL_LTC
	Symbol2Gmo[XRP2JPY_spt] = gomocoin.SYMBOL_XRP
	Symbol2Gmo[BTC2JPY_mgn] = gomocoin.SYMBOL_BTC_JPY
	Symbol2Gmo[ETH2JPY_mgn] = gomocoin.SYMBOL_ETH_JPY
	Symbol2Gmo[BCH2JPY_mgn] = gomocoin.SYMBOL_BCH_JPY
	Symbol2Gmo[LTC2JPY_mgn] = gomocoin.SYMBOL_LTC_JPY
	Symbol2Gmo[XRP2JPY_mgn] = gomocoin.SYMBOL_XRP_JPY

	GmoSymbol2Symbol = make(map[string]string)
	for symbol, gmo_symbol := range Symbol2Gmo {
		GmoSymbol2Symbol[gmo_symbol] = symbol
	}

	Mode2Gmo = make(map[string]string)
	Mode2Gmo[BTC2JPY_spt] = MODE_spot
	Mode2Gmo[ETH2JPY_spt] = MODE_spot
	Mode2Gmo[BCH2JPY_spt] = MODE_spot
	Mode2Gmo[LTC2JPY_spt] = MODE_spot
	Mode2Gmo[XRP2JPY_spt] = MODE_spot
	Mode2Gmo[BTC2JPY_mgn] = MODE_margin
	Mode2Gmo[ETH2JPY_mgn] = MODE_margin
	Mode2Gmo[BCH2JPY_mgn] = MODE_margin
	Mode2Gmo[LTC2JPY_mgn] = MODE_margin
	Mode2Gmo[XRP2JPY_mgn] = MODE_margin
}

func getGmoKey(name string) (string, error) {
	key, ok := Symbol2Gmo[name]
	if !ok {
		return "", gmoErrorf("cannot support '%s'", name)
	}

	return key, nil
}

type GmoConf struct {
	ApiKey string
	SecretKey string
}

func openGmo(conf *GmoConf, ctx context.Context) (*GmoHandler, error) {
	var key string
	var secret string
	if conf != nil {
		key = conf.ApiKey
		secret = conf.SecretKey
	}

	shop, err := gomocoin.NewGoMOcoin(key, secret, ctx)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}
	return &GmoHandler{
		shop: shop,
	}, nil
}

type GmoHandler struct {
	shop *gomocoin.GoMOcoin
}

func (self *GmoHandler) GetRate() (map[string]Rate, error) {
	rates, err := self.shop.GetRate()
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}

	i_rates := make(map[string]Rate)
	for key, val := range rates {
		i_rates[key] = val
	}
	return i_rates, nil
}

func (self *GmoHandler) GetPositions(symbol string) ([]Position, error) {
	key, err := getGmoKey(symbol)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}

	if isMargin(symbol) {
		return self.getMarginPositions(key)
	}
	return self.getSpotPositions(key)
}

func (self *GmoHandler) getMarginPositions(key string) ([]Position, error) {
	poss, err := self.shop.GetPositions(key)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}

	i_poss := []Position{}
	for _, pos := range poss {
		i_poss = append(i_poss, pos)
	}
	return i_poss, nil
}

func (self *GmoHandler) getSpotPositions(key string) ([]Position, error) {
	as, err := self.shop.GetAsset()
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}

	var no_fix_val float64
	for _, a := range as {
		if a.Symbol() != key {
			continue
		}

		no_fix_val = a.Available()
	}
	if no_fix_val <= float64(0) {
		return []Position{}, nil
	}

	pos := []Position{}
	fixes, err := self.shop.GetFixes(key)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}
	for _, fix := range fixes {
		if fix.OrderType() != TYPE_BUY {
			continue
		}

		if no_fix_val < float64(fix.Price() * fix.Size()) {
			break
		}
		pos = append(pos, fix)
		no_fix_val -= float64(fix.Price() * fix.Size())
	}

	if no_fix_val <= float64(0) {
		return pos, nil
	}

	rates, err := self.shop.GetRate()
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}
	rate, ok := rates[key]
	if !ok {
		return nil, gmoErrorf("cannot get rate.")
	}

	price := rate.Ask()
	pos = append(pos, &GmoSpotPosition{
		id: fmt.Sprintf("%v", time.Now().Unix()),
		symbol: key,
		size: no_fix_val,
		price: price,
		o_type: TYPE_BUY,
	})
	return pos, nil
}

func (self *GmoHandler) GetFixes(symbol string) ([]Fix, error) {
	key, err := getGmoKey(symbol)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}

	if isMargin(symbol) {
		return self.getMarginFixes(key)
	}
	return self.getSpotFixes(key)
}

func (self *GmoHandler) getMarginFixes(key string) ([]Fix, error) {
	fixes, err := self.shop.GetFixes(key)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}

	i_fixes := []Fix{}
	for _, fix := range fixes {
		i_fixes = append(i_fixes, fix)
	}
	return i_fixes, nil
}

func (self *GmoHandler) getSpotFixes(key string) ([]Fix, error) {
	return nil, gmoErrorf("not yet")
}

func (self *GmoHandler) Order(o_type string, symbol string,
							size float64, is_stream bool, price float64) error {
	key, err := getGmoKey(symbol)
	if err != nil {
		return gmoErrorf("%s", err)
	}

	if is_stream {
		if err := self.shop.MarketOrder(o_type, key, size); err != nil {
			return gmoErrorf("%s", err)
		}
		return nil
	}
	if isMargin(symbol) {
		if o_type != TYPE_BUY {
			return gmoErrorf("cannot operation '%s'", o_type)
		}
	}
	if err := self.shop.LimitOrder(o_type, key, size, price); err != nil {
		return gmoErrorf("%s", err)
	}
	return nil
}

func (self *GmoHandler) OrderFix(pos Position,
										is_stream bool, price float64) error {
	symbol, ok := GmoSymbol2Symbol[pos.Symbol()]
	if !ok {
		return gmoErrorf("undefined symbol '%s'", pos.Symbol())
	}

	if isMargin(symbol) {
		g_pos, ok := pos.(*gomocoin.Position)
		if !ok {
			return gmoErrorf("unkown type at this store.")
		}
		if is_stream {
			if err := self.shop.MarketOrderFix(g_pos); err != nil {
				return gmoErrorf("%s", err)
			}
			return nil
		}
		if err := self.shop.LimitOrderFix(g_pos, price); err != nil {
			return gmoErrorf("%s", err)
		}
		return nil
	}

	if pos.OrderType() != TYPE_BUY {
		return gmoErrorf("cannot fix operation '%s'", pos.OrderType())
	}
	if is_stream {
		if err := self.shop.MarketOrder(TYPE_SELL, pos.Symbol(), pos.Size()); err != nil {
			return gmoErrorf("%s", err)
		}
		return nil
	}
	if err := self.shop.LimitOrder(TYPE_SELL, pos.Symbol(), pos.Size(), price); err != nil {
		return gmoErrorf("%s", err)
	}
	return nil
}

func (self *GmoHandler) Release() error {
	if err := self.shop.Close(); err != nil {
		return gmoErrorf("%s", err)
	}
	return nil
}

type GmoSpotPosition struct {
	id     string
	symbol string
	size   float64
	price  float64
	o_type string
}

func (self *GmoSpotPosition) Id() string {
	return self.id
}

func (self *GmoSpotPosition) Symbol() string {
	return self.symbol
}

func (self *GmoSpotPosition) Size() float64 {
	return self.size
}

func (self *GmoSpotPosition) Price() float64 {
	return self.price
}

func (self *GmoSpotPosition) OrderType() string {
	return self.o_type
}

type GmoSpotFix struct {
/*
	Id()        string
	Symbol()    string
	OrderType() string
	Size()      float64
	Price()     float64
	Yield()     (float64, error)
	Date()      (time.Time, error)
*/
}
