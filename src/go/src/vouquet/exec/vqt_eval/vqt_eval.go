package main

import (
	"os"
	"fmt"
	"flag"
	"sync"
	"time"
	"context"
	"strconv"

	_ "time/tzdata"
)

import (
	"vouquet/soil"
)

import (
	"github.com/vouquet/florist"
)

var (
	Cpath    string

	Start  time.Time
	End    time.Time

	Symbol string
	Soil   string
	Size   float64

	Detail     bool
	VeryDetail bool
)

type logger struct {}

func (self *logger) WriteMsg(s string, msg ...interface{}) {
	if VeryDetail || Detail {
		fmt.Fprintf(os.Stdout, "[STDOUT] " + s + "\n" , msg...)
	}
}

func (self *logger) WriteErr(s string, msg ...interface{}) {
	if VeryDetail || Detail {
		fmt.Fprintf(os.Stderr, "[ERROR]" + s + "\n" , msg...)
	}
}

func (self *logger) WriteDebug(s string, msg ...interface{}) {
	if VeryDetail {
		fmt.Fprintf(os.Stdout, "[DEBUG] " + s + "\n" , msg...)
	}
}

func eval() error {
	before_data := Start.Add(-20 * time.Minute)

	log := new(logger)
	ctx := context.Background()
	r, err := soil.OpenRegistry(Cpath, ctx, log)
	if err != nil {
		return err
	}
	defer r.Close()

	status, err := r.GetStatus(Soil, Symbol, before_data, Start)
	if err != nil {
		return err
	}

	fls := make(map[string]florist.Florist)
	chan_list := make(map[string]chan *soil.State)
	for _, name := range florist.MEMBERS {

		p := soil.NewTestPlanter(Soil, Symbol, log)
		fl, err := florist.NewFlorist(name, p, status, log)
		if err != nil {
			return err
		}

		fls[name] = fl
		chan_list[name] = make(chan *soil.State)
	}

	var head time.Time
	var tail time.Time

	go func() {
		t_status, err := r.GetStatus(Soil, Symbol, Start, End)
		if err != nil {
			return
		}
		log.WriteMsg("read size: %v", len(t_status))

		head = t_status[0].Date()
		tail = t_status[len(t_status)-1].Date()

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

			if err := fl.Run(ctx, Size, chan_list[name]); err != nil {
				log.WriteErr("cannot run %s, %s", name, err)
			}
		}(name, fl)
	}
	wg.Wait()

	fmt.Printf("++++++++++++++++++++++++++++++++\n")
	fmt.Printf("vqt_eval report\n")
	fmt.Printf("simulate date: '%s' -> '%s'\n", head, tail)
	fmt.Printf("++++++++++++++++++++++++++++++++\n")
	for _, name := range florist.MEMBERS {
		fl, ok := fls[name]
		if !ok {
			fmt.Printf("%s : was not run test.", name)
			continue
		}

		fmt.Printf("%s : win: %f, order count: %v, win average: %v\n",
				name, fl.Yield(), fl.HarvestCnt(), fl.Yield() / float64(fl.HarvestCnt()))
	}
	fmt.Printf("++++++++++++++++++++++++++++++++\n")

	return nil
}

func die(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
	os.Exit(1)
}

func init() {
	var very_detail bool
	var detail bool
	var c_path string
	flag.StringVar(&c_path, "c", "./vouquet.conf", "config path.")
	flag.BoolVar(&detail, "v", false, "display detail.")
	flag.BoolVar(&very_detail, "vv", false, "display very very detail.")
	flag.Parse()

	if flag.NArg() < 5 {
		die("usage : vqt_eval [-c <config path>] [-v|-vv] <start-date(yyyy/mm/dd)> <end-date(yyyy/mm/dd)> <SYMBOL> <SOIL> <SIZE>")
	}
	if flag.NFlag() < 0 {
		die("usage : vqt_eval [-c <config path>] [-v|-vv] <start-date(yyyy/mm/dd)> <end-date(yyyy/mm/dd)> <SYMBOL> <SOIL> <SIZE>")
	}

	st_s := flag.Arg(0)
	et_s := flag.Arg(1)
	symbol := flag.Arg(2)
	soil := flag.Arg(3)
	size_s := flag.Arg(4)

	size, err := strconv.ParseFloat(size_s, 64)
	if err != nil {
		die("cannot convert float64 from size: '%s'", err)
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		die("cannot load timezone: Asia/Tokyo")
	}
	st_base, err := time.ParseInLocation("2006/01/02", st_s, jst)
	if err != nil {
		die("cannot convert time from start-date. '%s'", err)
	}
	st := time.Date(st_base.Year(), st_base.Month(), st_base.Day(), 0, 0, 0, 0, st_base.Location())

	et_base, err := time.ParseInLocation("2006/01/02", et_s, jst)
	if err != nil {
		die("cannot convert time from end-date. '%s'", err)
	}
	et := time.Date(et_base.Year(), et_base.Month(), et_base.Day(), 23, 59, 59, 999999999, et_base.Location())

	if c_path == "" {
		die("cannot set empty value of config path.")
	}
	if symbol == "" {
		die("cannot set empty value of symbol.")
	}
	if soil == "" {
		die("cannot set empty value of soil.")
	}

	Detail = detail
	VeryDetail = very_detail
	Cpath = c_path
	Symbol = symbol
	Soil = soil
	Size = size

	Start = st
	End = et
}

func main() {
	if err := eval(); err != nil {
		die("%s", err)
	}

}

