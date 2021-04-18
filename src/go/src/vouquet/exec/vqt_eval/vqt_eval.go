package main

import (
	"os"
	"fmt"
	"flag"
	"sort"
	"time"
	"context"
	"strconv"

	_ "time/tzdata"
)

import (
	"vouquet/farm"
	"vouquet/vouquet"
)

const (
	SELF_NAME string = "vqt_eval"
	USAGE string = "[-version] [-c <config path>] [-v|-vv] <start-date(yyyy/mm/dd)> <end-date(yyyy/mm/dd)> <SEED> <SOIL> <SIZE>"
)

var (
	Version string

	Cpath    string

	Start  time.Time
	End    time.Time

	Seed string
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
	fmt.Fprintf(os.Stderr, "[ERROR]" + s + "\n" , msg...)
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

	cfg, err := farm.LoadConfig(Cpath)
	if err != nil {
		return err
	}
	r, err := farm.OpenRegistry(cfg, ctx, log)
	if err != nil {
		return err
	}
	defer r.Close()

	status, err := r.GetStatus(Soil, Seed, before_data, Start)
	if err != nil {
		return err
	}

	fls := make(map[string]*vouquet.Florist)
	pls := make(map[string]farm.Planter)
	for _, name := range vouquet.FLORIST_NAMES {

		p := farm.NewTestPlanter(Seed, log)
		fl, err := vouquet.NewFlorist(name, p, status, log)
		if err != nil {
			return err
		}
		fl.SetSize(Size)

		fls[name] = fl
		pls[name] = p
	}

	var head time.Time
	var tail time.Time

	t_status, err := r.GetStatus(Soil, Seed, Start, End)
	if err != nil {
		return err
	}
	log.WriteMsg("read size: %v", len(t_status))

	if len(t_status) < 1 {
		return fmt.Errorf("This range havn't rate data.")
	}

	head = t_status[0].Date()
	tail = t_status[len(t_status)-1].Date()

	for _, t_state := range t_status {
		for _, p := range pls {
			tp, ok := p.(*farm.TestPlanter)
			if !ok {
				log.WriteErr("cannot convert test planter")
				continue
			}

			tp.SetState(t_state)
		}
		for _, fl := range fls {
			fl.Action(t_state)
		}
	}

	fmt.Printf("++++++++++++++++++++++++++++++++\n")
	fmt.Printf("vqt_eval %s report\n", Version)
	fmt.Printf("simulate date: '%s' -> '%s'\n", head, tail)
	fmt.Printf("++++++++++++++++++++++++++++++++\n")

	idx := []string{}
	for _, name := range vouquet.FLORIST_NAMES {
		idx = append(idx, name)
	}
	sort.SliceStable(idx, func(i, j int) bool { return idx[i] < idx[j] })
	for _, name := range idx {
		pl, ok := pls[name]
		if !ok {
			fmt.Printf("%s : was not run test.", name)
			continue
		}

		fmt.Printf("*** [%s] \n", name)
		fmt.Printf("    win: %.3f, count: %v, average: %.3f\n",
						pl.Win(), pl.WinCnt(), pl.Win() / float64(pl.WinCnt()))
		fmt.Printf("    lose: %.3f, count: %v, average: %.3f\n",
						pl.Lose(), pl.LoseCnt(), pl.Lose() / float64(pl.LoseCnt()))
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
	var see_version bool
	flag.StringVar(&c_path, "c", "./vouquet.conf", "config path.")
	flag.BoolVar(&detail, "v", false, "display detail.")
	flag.BoolVar(&very_detail, "vv", false, "display very very detail.")
	flag.BoolVar(&see_version, "version", false, "display version.")
	flag.Parse()

	if see_version {
		fmt.Printf("Version: %s %s\n", SELF_NAME, Version)
		os.Exit(0)
	}

	if flag.NArg() < 5 {
		die("usage : %s %s", SELF_NAME, USAGE)
	}
	if flag.NFlag() < 0 {
		die("usage : %s %s", SELF_NAME, USAGE)
	}

	st_s := flag.Arg(0)
	et_s := flag.Arg(1)
	seed := flag.Arg(2)
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
	if seed == "" {
		die("cannot set empty value of seed.")
	}
	if soil == "" {
		die("cannot set empty value of soil.")
	}

	Detail = detail
	VeryDetail = very_detail
	Cpath = c_path
	Seed = seed
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
