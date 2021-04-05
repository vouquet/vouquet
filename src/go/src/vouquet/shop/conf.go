package shop

type Conf interface {
	Gmo() *GmoConf
	Bitflyer() *BitflyerConf
}

type ConfBase struct {
	GMO      *GmoConf `toml:GMO`
	BITFLYER *BitflyerConf `toml:Bitflyer`
}

func (self *ConfBase) Gmo() *GmoConf {
	return self.GMO
}

func (self *ConfBase) Bitflyer() *BitflyerConf {
	return self.BITFLYER
}
