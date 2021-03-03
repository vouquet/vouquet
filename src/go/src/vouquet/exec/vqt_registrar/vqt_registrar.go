package main

import (
	"os"
	"fmt"
	"sync"
	"time"
	"context"
)

import (
	"vouquet/soil"
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
	r, err := soil.OpenRegistry("./.vouquet", ctx, log)
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
			for {
				time.Sleep(time.Second)

				ss, err := t.Status()
				if err != nil {
					log.WriteErr("%s", err)
					continue
				}

				if err := r.Record(ss); err != nil {
					log.WriteErr("%s", err)
					continue
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

func main() {
	if err := registrar(); err != nil {
		die("%s", err)
	}

}
