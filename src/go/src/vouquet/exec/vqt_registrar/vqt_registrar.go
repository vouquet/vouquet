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

func registrar() error {
	log := new(logger)

	ctx := context.Background()
	r, err := farm.OpenRegistry(Cpath, ctx, log)
	if err != nil {
		return err
	}
	defer r.Close()

	wg := new(sync.WaitGroup)
	for _, s := range farm.SOIL_ALL {
		t, err := farm.NewThemograpy(Cpath, s, ctx)
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

	log.WriteMsg("Start %s %s", SELF_NAME, Version)
	wg.Wait()
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
