package base

import (
	"vouquet/farm"
)

type Logger interface{
	WriteErr(string, ...interface{})
	WriteMsg(string, ...interface{})
	WriteDebug(string, ...interface{})
}

type Florist interface {
	Release() error

	SetPlanter(farm.Planter)
	SetLogger(Logger)
	SetSize(float64)
	Size() float64

	Init([]*farm.State) error
	Action(*farm.State) error
}

type FloristBase struct {
	planter farm.Planter
	size    float64

	log     Logger
}

func (self *FloristBase) Planter() farm.Planter {
	return self.planter
}

func (self *FloristBase) SetPlanter(p farm.Planter) {
	self.planter = p
}

func (self *FloristBase) Logger() Logger {
	return self.log
}

func (self *FloristBase) SetLogger(log Logger) {
	self.log = log
}

func (self *FloristBase) Size() float64 {
	return self.size
}

func (self *FloristBase) SetSize(size float64) {
	self.size = size
}

func (self *FloristBase) Release() error {
	return self.planter.Release()
}
