package shop

import "log"

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

	mapped  map[string]struct{}
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
		log.Println("symbol:", a.Symbol())
		log.Println("val", a.Available(), a.Amount())
		if a.Symbol() != key {
			continue
		}

		no_fix_val = a.Available()
	}
	log.Println("no fix val:", no_fix_val)
	if no_fix_val <= float64(0) {
		return []Position{}, nil
	}

	pos := []Position{}
	fixes, err := self.shop.GetFixes(key)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}
	log.Println("fixes:", len(fixes))
	for _, fix := range fixes {
		log.Println("fix:", fix)
		if fix.OrderType() != TYPE_BUY {
			continue
		}

		if no_fix_val < fix.Size() {
			log.Println("break detected", no_fix_val, fix.Size())
			break
		}

		pos = append(pos, fix)
		log.Println("fix size", float64Sub(fix.Size(), fix.Size()))
		log.Println("no_fix_val", float64Sub(no_fix_val, no_fix_val))
		no_fix_val = float64Sub(no_fix_val, fix.Size())
		log.Println("mathed no_fix_val", no_fix_val)
	}

	if no_fix_val <= float64(0) {
		log.Println("possize", len(pos))
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
	log.Println("possize(break", len(pos))
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
	fixes, err := self.shop.GetFixes(key)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}

	if self.mapped == nil {
		self.mapped = map[string]struct{}{}
		for _, fix := range fixes {
			self.mapped[fix.Id()] = struct{}{}
		}

		return []Fix{}, nil
	}

	sell_buf := []*gomocoin.Fix{}
	buy_buf := []*gomocoin.Fix{}
	for _, fix := range fixes {
		_, ok := self.mapped[fix.Id()]
		if ok {
			continue
		}
		switch fix.OrderType() {
		case TYPE_SELL:
			sell_buf = append(sell_buf, fix)
		case TYPE_BUY:
			buy_buf = append(buy_buf, fix)
		}
	}
	if len(buy_buf) < 1 {
		for _, s_fix := range sell_buf {
			self.mapped[s_fix.Id()] = struct{}{}
		}
		return []Fix{}, nil
	}

	ret_fixes := []Fix{}
	for _, s_fix := range sell_buf {
		for _, b_fix := range buy_buf {
			if b_fix.Size() != s_fix.Size() {
				continue
			}

			date, err := s_fix.Date()
			if err != nil {
				return nil, gmoErrorf("%s", err)
			}
			ret_fixes = append(ret_fixes, &GmoSpotFix{
				id: b_fix.Id() + s_fix.Id(),
				symbol: s_fix.Symbol(),
				o_type: TYPE_BUY,
				size: s_fix.Size(),
				price: s_fix.Price(),
				yield: (s_fix.Price() - b_fix.Price()) * s_fix.Size(),
				date: date,
			})
			self.mapped[b_fix.Id()] = struct{}{}
			self.mapped[s_fix.Id()] = struct{}{}
			break
		}
	}
	return ret_fixes, nil
}

func (self *GmoHandler) Order(o_type string, symbol string,
							size float64, is_stream bool, price float64) error {
	key, err := getGmoKey(symbol)
	if err != nil {
		return gmoErrorf("%s", err)
	}

	if !isMargin(symbol) {
		if o_type != TYPE_BUY {
			return gmoErrorf("cannot operation '%s'", o_type)
		}
	}
	if is_stream {
		if err := self.shop.MarketOrder(key, o_type, size); err != nil {
			return gmoErrorf("%s", err)
		}
		return nil
	}
	if err := self.shop.LimitOrder(key, o_type, size, price); err != nil {
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
		if err := self.shop.MarketOrder(pos.Symbol(), TYPE_SELL, pos.Size()); err != nil {
			return gmoErrorf("%s", err)
		}
		return nil
	}
	if err := self.shop.LimitOrder(pos.Symbol(), TYPE_SELL, pos.Size(), price); err != nil {
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
	id     string
	symbol string
	o_type string
	size   float64
	price  float64
	yield  float64
	date   time.Time
}

func (self *GmoSpotFix) Id() string {
	return self.id
}

func (self *GmoSpotFix) Symbol() string {
	return self.symbol
}

func (self *GmoSpotFix) OrderType() string {
	return self.o_type
}

func (self *GmoSpotFix) Size() float64 {
	return self.size
}

func (self *GmoSpotFix) Price() float64 {
	return self.price
}

func (self *GmoSpotFix) Yield() (float64, error) {
	return self.yield, nil
}

func (self *GmoSpotFix) Date() (time.Time, error) {
	return self.date, nil
}
