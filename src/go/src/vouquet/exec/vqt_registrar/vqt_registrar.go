package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"context"
)

import (
	"vouquet/lock"
	"vouquet/farm"
)

const (
	SELF_NAME string = "vqt_registrar"
	USAGE string = "[-version] [-c <config path>]"
)

var (
	Version string
	Cpath string
)

type logger struct {}

func (self *logger) WriteMsg(s string, msg ...interface{}) {
	tstr := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(os.Stdout, tstr + " " + s + "\n" , msg...)
}

func (self *logger) WriteErr(s string, msg ...interface{}) {
	tstr := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(os.Stderr, tstr + " " + s + "\n" , msg...)
}

func (self *logger) WriteDebug(s string, msg ...interface{}) {
	tstr := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(os.Stdout, tstr + " [DEBUG] " + s + "\n" , msg...)
}

type worker struct {
	soil_name   string
	c_path      string

	registry    *farm.Registry
	themograpy  *farm.Themography

	fail_cnt    int

	log         *logger
	ctx         context.Context
	mtx         *lock.TryMutex
}

func NewWorker(r *farm.Registry, soil_name string, c_path string, ctx context.Context, log *logger) *worker {
	return &worker{
		soil_name: soil_name,
		c_path: c_path,
		registry:r,

		fail_cnt: 60,

		log: log,
		ctx: ctx,
		mtx: lock.NewTryMutex(ctx),
	}
}

func (self *worker) Do() error {
	ok, err := self.mtx.TryLock()
	if err != nil {
		self.fail_cnt++
		if err == lock.ERR_CONTEXT_CANCEL {
			return nil
		}
		return err
	}
	if !ok {
		self.fail_cnt++
		return nil
	}
	defer self.mtx.Unlock()

	if self.themograpy == nil {
		if err := self.open(); err != nil {
			return err
		}
	}

	ss, err := self.themograpy.Status()
	if err != nil {
		self.fail_cnt++
		self.failedSleep()
		return err
	}
	self.fail_cnt--
	if err := self.registry.Record(ss); err != nil {
		return err
	}
	return nil
}

func (self *worker) failedSleep() {
	if self.fail_cnt <= 10 {
		return
	}
	if self.fail_cnt <= 60 {
		self.sleep(time.Second * 60)
		return
	}
	if self.fail_cnt <= (60 * 5) {
		self.sleep(time.Second * 60 * 5)
		return
	}
	if self.fail_cnt <= (60 * 15) {
		self.sleep(time.Second * 60 * 15)
		return
	}
	if self.fail_cnt <= (60 * 30) {
		self.sleep(time.Second * 60 * 30)
		return
	}
	if self.fail_cnt <= (60 * 60) {
		self.sleep(time.Second * 60 * 60)
		return
	}
	if self.fail_cnt <= (60 * 60 * 6) {
		self.sleep(time.Second * 60 * 60 * 6)
		return
	}
	self.sleep(time.Second * 60 * 60 * 12)
}

func (self *worker) sleep(size time.Duration) {
	t := time.NewTimer(size)
	defer t.Stop()

	self.log.WriteMsg("%s error detected. sleep %s", self.soil_name, size)
	defer self.log.WriteMsg("%s sleep (%s) done.", self.soil_name, size)
	select {
	case <- self.ctx.Done():
		return
	case <- t.C:
	}
}

func (self *worker) ThemograpyRelease() error {
	if err := self.mtx.Lock(); err !=nil {
		if err == lock.ERR_CONTEXT_CANCEL {
			return nil
		}
		return err
	}
	defer self.mtx.Unlock()

	if self.themograpy == nil {
		return nil
	}
	return self.themograpy.Release()
}

func (self *worker) open() error {
	t, err := farm.NewThemograpy(self.c_path, self.soil_name, self.ctx)
	if err != nil {
		return err
	}
	self.themograpy = t
	return nil
}

func registrar() error {
	log := new(logger)
	ctx, cancel := context.WithCancel(context.Background())

	r, err := farm.OpenRegistry(Cpath, ctx, log)
	if err != nil {
		return err
	}
	defer r.Close()

	wks := []*worker{}
	for _, s := range farm.SOIL_ALL {
		wks = append(wks, NewWorker(r, s, Cpath, ctx, log))
	}
	defer func() {
		for _, wk := range wks {
			wk.ThemograpyRelease()
		}
	}()

	log.WriteMsg("Start %s %s", SELF_NAME, Version)

	defer cancel()

	timer := time.NewTicker(time.Second)
	for {
		select {
		case <- ctx.Done():
			return nil
		case <- timer.C:
			go func() {
				for _, wk := range wks {
					go func(wk *worker) {
						if err := wk.Do(); err != nil {
							log.WriteErr("%s", err)
						}
					}(wk)
				}
			}()
		}
	}

	return nil
}

func die(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
	os.Exit(1)
}

func init() {
	var c_path string
	var see_version bool
	flag.StringVar(&c_path, "c", "./vouquet.conf", "config path.")
	flag.BoolVar(&see_version, "version", false, "display version.")
	flag.Parse()

	if see_version {
		fmt.Printf("Version: %s %s\n", SELF_NAME, Version)
		os.Exit(0)
	}

	if flag.NArg() < 0 {
		die("usage : %s %s", SELF_NAME, USAGE)
	}

	if c_path == "" {
		die("empty path")
	}

	Cpath = c_path
}

func main() {
	if err := registrar(); err != nil {
		die("%s", err)
	}

}
