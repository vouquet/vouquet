package soil

type Planter interface {
	SetSeed(string, float64, float64) error
	ShowSproutList() ([]*Sprout, error)
	Harvest(*Sprout) error
	Yield() float64
}

type Flowerpot struct {
	r      *Registry

	log    logger

	ctx    context.Context
	cancel context.CancelFunc
	mtx    *sync.Mutex
}

func NewFlowerpot(name string, symbol string c_path string, ctx context.Context, log logger) (*Flowerpots, error) {
}

func (self *Flowerpot) SetSeed(o_type string, size float64, rate float64) error {
}

func (self *Flowerpot) ShowSproutList() ([]*Sprout, error) {
}

func (self *Flowerpot) Harvest(sp *Sprout) error {
}

func (self *Flowerpot) Yield() float64 {
}

type Sprout struct {
}

type Tester struct {
}

func NewTester(symbol string, ctx context.Context, log logger) (*Tester, error) {
}

func (self *Tester) SetSeed(o_type string, size float64, rate float64) error {
}

func (self *Tester) ShowSproutList() ([]*Sprout, error) {
}

func (self *Tester) Harvest(sp *Sprout) error {
}

func (self *Tester) Yield() float64 {
}
