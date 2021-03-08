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
	"vouquet/soil"
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

func registrar() error {
	log := new(logger)

	ctx := context.Background()
	r, err := soil.OpenRegistry(Cpath, ctx, log)
	if err != nil {
		return err
	}
	defer r.Close()

	wg := new(sync.WaitGroup)
	for _, s := range soil.SOIL_ALL {
		t, err := soil.NewThemograpy(s, ctx)
		if err != nil {
			return err
		}
		defer t.Release()

		wg.Add(1)
		go func () {
			defer wg.Done()

			mtx := new(sync.Mutex)
			timer := time.NewTicker(time.Second)
			for {
				select {
				case <- ctx.Done():
					return
				case <- timer.C:
					go func() {
						mtx.Lock()
						defer mtx.Unlock()

						ss, err := t.Status()
						if err != nil {
							log.WriteErr("%s", err)
							return
						}

						if err := r.Record(ss); err != nil {
							log.WriteErr("%s", err)
							return
						}
					}()
				}
			}
		}()
	}
	wg.Wait()
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
		die("usage : vqt_registrar [-c <config path>]")
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
