package farm

import (
	"fmt"
	"sort"
	"sync"
	"time"
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

import (
	"vouquet/shop"
)

type State struct {
	ask  float64
	bid  float64

	date time.Time
}

func (self *State) Ask() float64 {
	return self.ask
}

func (self *State) Bid() float64 {
	return self.bid
}

func (self *State) Date() time.Time {
	return self.date
}

type Registry struct {
	db  *sql.DB

	log    logger

	ctx    context.Context
	cancel context.CancelFunc
	mtx    *sync.Mutex
}

func OpenRegistry(cfg *Config, ctx context.Context, log logger) (*Registry, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	c_ctx, cancel := context.WithCancel(ctx)

	opt := "?parseTime=true&loc=Local"
	db, err := sql.Open("mysql", cfg.sqlcred() + opt)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(c_ctx); err != nil {
		db.Close()
		return nil, err
	}

	self := &Registry{db:db, log:log, ctx:c_ctx, cancel:cancel, mtx:new(sync.Mutex)}
	if err := self.checktbl(); err != nil {
		self.Close()
		return nil, err
	}
	return self, nil
}

func (self *Registry) Close() error {
	self.lock()
	defer self.unlock()

	self.cancel()
	return self.db.Close()
}

func (self *Registry) Record(ss *Status) error {
	self.lock()
	defer self.unlock()

	for seed, rate := range ss.rates {
		err := self.do_sql_updateSymbol(ss.soil_name, seed, rate.Ask(), rate.Bid())
		if err != nil {
			self.log.WriteErr("Registry.Record: %s", err)
		}
	}
	return nil
}

func (self *Registry) GetStatus(soil string, seed string, st time.Time, et time.Time) ([]*State, error) {
	self.lock()
	defer self.unlock()

	key, err := shop.GetKey(soil, seed)
	if err != nil {
		return nil, err
	}

	return self.do_sql_getStatus(soil, key, st, et)
}

func (self *Registry) GetLastState(soil string, seed string) (*State, error) {
	self.lock()
	defer self.unlock()

	key, err := shop.GetKey(soil, seed)
	if err != nil {
		return nil, err
	}

	return self.do_sql_getLastState(soil, key)
}

func (self *Registry) checktbl() error {
	soils, err := self.do_sql_getTables()
	if err != nil {
		return err
	}

	soils_idx := make(map[string]interface{})
	for _, soil := range soils {
		soils_idx[soil] = nil
	}

	for _, soil := range SOIL_ALL {
		_, ok := soils_idx[soil]
		if ok {
			continue
		}

		if err := self.do_sql_createTable(soil); err != nil {
			return err
		}
	}
	return nil
}

func (self *Registry) do_sql_getTables() ([]string, error) {
	qstr := "SHOW TABLES;"

	rows, err := self.db.QueryContext(self.ctx, qstr)
	if err != nil {
		return nil, fmt.Errorf("do_sql_getTables: query: '%s', err: '%s'", qstr, err)
	}
	defer rows.Close()

	var soils []string
	for rows.Next() {
		var soil string
		if err := rows.Scan(&soil); err != nil {
			return nil, err
		}
		soils = append(soils, soil)
	}
	return soils, nil
}

func (self *Registry) do_sql_updateSymbol(soil string, seed string, ask float64, bid float64) error {
	base := "INSERT INTO %s (symbol, ask, bid) VALUES ('%s', %f, %f)"
	qstr := fmt.Sprintf(base, soil, seed, ask, bid)

	if _, err := self.db.ExecContext(self.ctx, qstr); err != nil {
		return fmt.Errorf("do_sql_updateSymbol: query: '%s', err: '%s'", qstr, err)
	}
	return nil
}

func (self *Registry) do_sql_getLastState(soil string, seed string) (*State, error) {
	base := "SELECT time, ask, bid FROM %s WHERE symbol = '%s' ORDER BY time DESC limit 1;"
	qstr := fmt.Sprintf(base, soil, seed)

	rows, err := self.db.QueryContext(self.ctx, qstr)
	if err != nil {
		return nil, fmt.Errorf("do_sql_getLastState: query: '%s', err: '%s'", qstr, err)
	}
	defer rows.Close()

	var state *State
	for rows.Next() {
		var t time.Time
		var ask float64
		var bid float64
		if err := rows.Scan(&t, &ask, &bid); err != nil {
			return nil, fmt.Errorf("do_sql_getLastState: cannot convert slqdata. : '%s'", err)
		}

		state = &State{ask: ask, bid: bid, date: t}
		break
	}
	if state == nil {
		return nil, fmt.Errorf("do_sql_getLastState: not found data.")
	}
	return state, nil
}

func (self *Registry) do_sql_getStatus(soil string, seed string, st time.Time, et time.Time) ([]*State, error) {
	base := "SELECT time, ask, bid FROM %s WHERE symbol = '%s' AND time BETWEEN '%s:00' AND '%s:59';"
	t_fmt := "2006-01-02 15:04"
	qstr := fmt.Sprintf(base, soil, seed, st.Format(t_fmt), et.Format(t_fmt))

	rows, err := self.db.QueryContext(self.ctx, qstr)
	if err != nil {
		return nil, fmt.Errorf("do_sql_getStatus: query: '%s', err: '%s'", qstr, err)
	}
	defer rows.Close()

	status := []*State{}
	for rows.Next() {
		var t time.Time
		var ask float64
		var bid float64
		if err := rows.Scan(&t, &ask, &bid); err != nil {
			self.log.WriteErr("Registry.do_sql_getStatus: cannot convert sqldata. :'%s'", err)
			continue
		}

		state := &State{ask: ask, bid: bid, date: t}
		status = append(status, state)

	}
	sort.SliceStable(status, func(i, j int) bool { return status[i].date.Before(status[j].date) })
	return status, nil
}

func (self *Registry) do_sql_createTable(name string) error {
	base := `CREATE TABLE %s (
symbol VARCHAR(11) NOT NULL,
time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
ask DOUBLE,
bid DOUBLE,
PRIMARY KEY(symbol, time));`

	qstr := fmt.Sprintf(base, name)
	if _, err := self.db.ExecContext(self.ctx, qstr); err != nil {
		return fmt.Errorf("do_sql_createTable: query: '%s', err: '%s'", qstr, err)
	}
	return nil
}

func (self *Registry) lock() {
	self.mtx.Lock()
}

func (self *Registry) unlock() {
	self.mtx.Unlock()
}
