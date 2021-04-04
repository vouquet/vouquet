package shop

type Conf interface {
	Gmo() *GmoConf
}

type ConfBase struct {
	GMO *GmoConf `toml:GMO`
}

func (self *ConfBase) Gmo() *GmoConf {
	return self.GMO
}
