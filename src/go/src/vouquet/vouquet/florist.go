package vouquet

import (
	"fmt"
	"sync"
	"context"
)

import (
	"vouquet/farm"
	"vouquet/base"

	"github.com/vouquet/florist"
)

var (
	FLORIST_NAMES []string
)

type Florist struct {
	original base.Florist
	mtx      *sync.Mutex
}

func NewFlorist(name string, p farm.Planter, state []*farm.State, log base.Logger) (*Florist, error) {
	f, ok := florist.MEMBERS[name]
	if !ok {
		return nil, fmt.Errorf("undefined member : '%s'", name)
	}
	o := f()
	o.SetPlanter(p)
	o.SetLogger(log)

	if err := o.Init(state); err != nil {
		return nil, err
	}
	return newFlorist(o), nil
}

func newFlorist(original base.Florist) *Florist {
	return &Florist{
		original: original,
		mtx: new(sync.Mutex),
	}
}

func (self *Florist) Run(ctx context.Context, st_ch <- chan *farm.State) error {
	if self.original.Size() == float64(0) {
		return fmt.Errorf("cannot use value zero of size.")
	}

	for {
		select {
		case <- ctx.Done():
			return nil
		case s, ok := <- st_ch:
			if !ok {
				return nil
			}
			if s == nil{
				return nil
			}

			func(s farm.State) {
				self.mtx.Lock()
				defer self.mtx.Unlock()

				self.original.Action(&s)
			}(*s)
		}
	}
}

func (self *Florist) SetSize(size float64) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	self.original.SetSize(size)
}

func (self *Florist) Action(st *farm.State) error {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	if self.original.Size() == float64(0) {
		return fmt.Errorf("cannot use value zero of size.")
	}
	return self.original.Action(st)
}

func init() {
	FLORIST_NAMES = []string{}
	for name, _ := range florist.MEMBERS {
		FLORIST_NAMES = append(FLORIST_NAMES, name)
	}
}
