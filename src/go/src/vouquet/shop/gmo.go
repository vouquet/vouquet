package shop

import (
	"fmt"
	"context"
)

import (
	"github.com/vouquet/go-gmo-coin/gomocoin"
)

var (
	Symbol2Gmo map[string]string
	Mode2Gmo   map[string]string
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

func (self *GmoHandler) GetFixes(symbol string) ([]Fix, error) {
	key, err := getGmoKey(symbol)
	if err != nil {
		return nil, gmoErrorf("%s", err)
	}

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

func (self *GmoHandler) Order(o_type string, symbol string,
							size float64, is_stream bool, price float64) error {
	key, err := getGmoKey(symbol)
	if err != nil {
		return gmoErrorf("%s", err)
	}

	if !is_stream {
		return gmoErrorf("not suport yet")
	}
	return self.shop.OrderStreamIn(o_type, key, size)
}

func (self *GmoHandler) OrderFix(pos Position,
										is_stream bool, price float64) error {
	g_pos, ok := pos.(*gomocoin.Position)
	if !ok {
		return gmoErrorf("unkown type at this store.")
	}
	if !is_stream {
		return gmoErrorf("not suport yet")
	}
	return self.shop.OrderStreamOut(g_pos)
}

func (self *GmoHandler) Release() error {
	if err := self.shop.Close(); err != nil {
		return gmoErrorf("%s", err)
	}
	return nil
}
