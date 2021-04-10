package shop

type Conf interface {
	Gmo() *GmoConf
	Bitflyer() *BitflyerConf
	Coincheck() *CoincheckConf
}

type ConfBase struct {
	GMO       *GmoConf       `toml:GMO`
	BITFLYER  *BitflyerConf  `toml:Bitflyer`
	COINCHECK *CoincheckConf `toml:Coincheck`
}

func (self *ConfBase) Gmo() *GmoConf {
	return self.GMO
}

func (self *ConfBase) Bitflyer() *BitflyerConf {
	return self.BITFLYER
}

func (self *ConfBase) Coincheck() *CoincheckConf {
	return self.COINCHECK
}
