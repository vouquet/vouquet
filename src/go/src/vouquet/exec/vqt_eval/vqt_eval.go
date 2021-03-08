package main

import (
	"os"
	"fmt"
	"flag"
	"sync"
	"time"
	"context"
	"strconv"
)

import (
	"vouquet/soil"
	"vouquet/seed"
)

import (
	"github.com/vouquet/florist"
)

const (
	TEST_SIZE = 0.001
)

var (
	Cpath    string
	Detail   bool
	TestDays int
)

type logger struct {}

func (self *logger) WriteMsg(s string, msg ...interface{}) {
	if Detail {
		fmt.Fprintf(os.Stdout, "[STDOUT] " + s + "\n" , msg...)
	}
}

func (self *logger) WriteErr(s string, msg ...interface{}) {
	if Detail {
		fmt.Fprintf(os.Stderr, "[ERROR]" + s + "\n" , msg...)
	}
}

func eval() error {
	now := time.Now()
	start := now.AddDate(0, 0, int(TestDays * -1))
	before_data := start.Add(-20 * time.Minute)

	log := new(logger)
	ctx := context.Background()
	r, err := soil.OpenRegistry(Cpath, ctx, log)
	if err != nil {
		return err
	}
	defer r.Close()

	status, err := r.GetStatus(soil.SOIL_GMO, seed.SYMBOL_BTC_REVERAGE, before_data, start)
	if err != nil {
		return err
	}

	fls := make(map[string]florist.Florist)
	chan_list := make(map[string]chan *soil.State)
	for _, name := range florist.MEMBERS {

		p := soil.NewTestPlanter(name, seed.SYMBOL_BTC_REVERAGE, log)
		fl, err := florist.NewFlorist(name, p, status, log)
		if err != nil {
			return err
		}

		fls[name] = fl
		chan_list[name] = make(chan *soil.State)
	}

	go func() {
		t_status, err := r.GetStatus(soil.SOIL_GMO, seed.SYMBOL_BTC_REVERAGE, start, now)
		if err != nil {
			return
		}
		log.WriteMsg("read size: %v", len(t_status))

		for _, t_state := range t_status {
			for _, s_chan := range chan_list {
				s_chan <- t_state
			}
		}

		for _, s_chan := range chan_list {
			close(s_chan)
		}
	}()

	wg := new(sync.WaitGroup)
	for name, fl := range fls {
		wg.Add(1)

		go func(name string, fl florist.Florist) {
			defer wg.Done()

			if err := fl.Run(ctx, TEST_SIZE, chan_list[name]); err != nil {
				log.WriteErr("cannot run %s, %s", name, err)
			}
		}(name, fl)
	}
	wg.Wait()

	fmt.Printf("++++++++++++++++++++++++++++++++\n")
	fmt.Printf("vqt_eval report\n")
	fmt.Printf("test days: %v ('%s' -> '%s')\n", TestDays, start, now)
	fmt.Printf("++++++++++++++++++++++++++++++++\n")
	for name, fl := range fls {
		fmt.Printf("%s : win: %f, order count: %v\n", name, fl.Yield(), fl.HarvestCnt())
	}
	fmt.Printf("++++++++++++++++++++++++++++++++\n")

	return nil
}

func die(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
	os.Exit(1)
}

func init() {
	var detail bool
	var c_path string
	flag.StringVar(&c_path, "c", "./vouquet.conf", "config path.")
	flag.BoolVar(&detail, "v", false, "display detail.")
	flag.Parse()

	if flag.NArg() < 1 {
		die("usage : vqt_test [-c <config path>] [-v] <test days>")
	}
	if flag.NFlag() < 0 {
		die("usage : vqt_test [-c <config path>] [-v] <test days>")
	}

	test_days, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		die("err: '%s", err)
	}

	if test_days < 1 {
		die("cannot set size of test days less than 1.")
	}
	if c_path == "" {
		die("empty path")
	}

	Detail = detail
	TestDays = test_days
	Cpath = c_path
}

func main() {
	if err := eval(); err != nil {
		die("%s", err)
	}

}

