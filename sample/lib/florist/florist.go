package florist

import (
	"vouquet/farm"
	"vouquet/base"
)

var (
	MEMBERS map[string]func() base.Florist
)

func init() {
	MEMBERS = make(map[string]func() base.Florist)
	MEMBERS["Alex"] = NewAlex
	MEMBERS["Johns"] = NewJohns
}

type Alex struct {
	base.FloristBase

	//Your super buffer, struct, etc...
}

func NewAlice() base.Florist {
	return &Alex{}
}

func (self *Alex) Init(ss []*farm.State) error {
	//Your super logic.
	return nil
}

func (self *Alex) Action(s *farm.State) error {
	//Your super logic.
	return nil
}

type Johns struct {
	base.FloristBase

	//Your super buffer, struct, etc...
}

func NewJohns() base.Florist {
	return &Johns{}
}

func (self *Johns) Init(ss []*farm.State) error {
	//Your super logic.
	return nil
}

func (self *Johns) Action(s *farm.State) error {
	//Your super logic.
	return nil
}
