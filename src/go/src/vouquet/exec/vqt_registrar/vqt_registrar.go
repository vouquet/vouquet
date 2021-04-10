package main

import (
	"os"
	"fmt"
	"flag"
	"sync"
	"time"
	"context"
)

import (
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

	ctx         context.Context
	mtx         *sync.Mutex
}

func NewWorker(r *farm.Registry, soil_name string, c_path string, ctx context.Context) *worker {
	return &worker{
		soil_name: soil_name,
		c_path: c_path,
		registry:r,

		ctx: ctx,
		mtx: new(sync.Mutex),
	}
}

func (self *worker) Do() error {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	if self.themograpy == nil {
		if err := self.open(); err != nil {
			return err
		}
	}

	ss, err := self.themograpy.Status()
	if err != nil {
		return err
	}
	if err := self.registry.Record(ss); err != nil {
		return err
	}
	return nil
}

func (self *worker) ThemograpyRelease() error {
	self.mtx.Lock()
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
		wks = append(wks, NewWorker(r, s, Cpath, ctx))
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
