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
	p := soil.NewTestPlanter(seed.SYMBOL_BTC_REVERAGE)

	asks, bids, err := r.GetStatusPerMinute(soil.SOIL_GMO, seed.SYMBOL_BTC_REVERAGE, before_data, start)
	if err != nil {
		return err
	}

	fls := make(map[string]florist.Florist)
	chan_list := make(map[string]chan *florist.StatusGroup)
	for _, name := range florist.MEMBERS {
		fl, err := florist.NewFlorist(name, p, asks, bids)
		if err != nil {
			return err
		}

		fls[name] = fl
		chan_list[name] = make(chan *florist.StatusGroup)
	}

	go func() {
		t_asks, t_bids, err := r.GetStatusPerMinute(soil.SOIL_GMO, seed.SYMBOL_BTC_REVERAGE, start, now)
		if err != nil {
			return
		}
		log.WriteMsg("read size: %v", len(t_asks))

		for i, ask := range t_asks {
			sg, err := florist.NewStatusGroup(ask, t_bids[i])
			if err != nil {
				log.WriteErr("cannot convert status group. '%s'", err)
				continue
			}

			for _, sg_chan := range chan_list {
				sg_chan <- sg
			}
		}

		for _, sg_chan := range chan_list {
			close(sg_chan)
		}
	}()

	wg := new(sync.WaitGroup)
	for name, fl := range fls {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := fl.Run(ctx, chan_list[name]); err != nil {
				log.WriteErr("cannot run %s, %s", name, err)
			}
		}()
	}
	wg.Wait()

	fmt.Printf("++++++++++++++++++++++++++++++++\n")
	fmt.Printf("vqt_eval report\n")
	fmt.Printf("test days: %v ('%s' -> '%s')\n", TestDays, start, now)
	fmt.Printf("++++++++++++++++++++++++++++++++\n")
	for name, fl := range fls {
		fmt.Printf("%s : win: %f\n", name, fl.Yield())
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

