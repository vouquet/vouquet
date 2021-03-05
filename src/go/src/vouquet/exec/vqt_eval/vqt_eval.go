package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"context"
)

import (
	"vouquet/soil"
	"vouquet/seed"
)

var (
	Cpath string
)

type logger struct {}

func (self *logger) WriteMsg(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stdout, s + "\n" , msg...)
}

func (self *logger) WriteErr(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
}

func eval() error {
	log := new(logger)

	ctx := context.Background()
	r, err := soil.OpenRegistry(Cpath, ctx, log)
	if err != nil {
		return err
	}
	defer r.Close()

	p := soil.NewTestPlanter(seed.SYMBOL_BTC)

	asks, bids, err := r.GetStatusPerMinute(soil.SOIL_GMO, seed.SYMBOL_BTC, time.Now().Add(-60 * time.Minute), time.Now())
	if err != nil {
		return err
	}
	for i, ask := range asks {
		p.SetSeed(soil.TYPE_BUY, 0.01, ask.Close())
		sps, _ := p.ShowSproutList()
		for _, sp := range sps {
			p.Harvest(sp, bids[i].Close())
		}
	}
	log.WriteMsg("yield : %f", p.Yield())

	return nil
}

func die(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
	os.Exit(1)
}

func init() {
	var c_path string
	flag.StringVar(&c_path, "c", "./vouquet.conf", "config path.")
	flag.Parse()

	if flag.NArg() < 0 {
		die("usage : vqt_test [-c <config path>]")
	}

	if c_path == "" {
		die("empty path")
	}

	Cpath = c_path
}

func main() {
	if err := eval(); err != nil {
		die("%s", err)
	}

}

