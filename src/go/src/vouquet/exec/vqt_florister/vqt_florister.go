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
	"vouquet/vouquet"
)

const (
	SELF_NAME string = "vqt_florister"
	USAGE string = "[-version] [-c <config path>]"
)

var (
	Version string
	Cpath  string
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

type Worker struct {
	fl     *vouquet.Florist
	st_ch  chan *farm.State

	work   *farm.Work

	before_state *farm.State

	log    *logger
	mtx    *lock.TryMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewWorker(b_ctx context.Context, log *logger, cfg *farm.Config,
						wk *farm.Work, status []*farm.State) (*Worker, error) {
	ctx, cancel := context.WithCancel(b_ctx)

	var pl farm.Planter
	if wk.PrdMode {
		var err error
		pl, err = farm.NewFlowerpot(wk.Soil, wk.Seed, cfg, ctx, log)
		if err != nil {
			return nil, err
		}
		log.WriteMsg("Load **Prd** worker %s, soil: %s, seed: %s, size: %f",
								wk.Florist, wk.Soil, wk.Seed, wk.Size)
	} else {
		pl = farm.NewTestPlanter(wk.Seed, log)
		log.WriteMsg("Load Demo worker %s, soil: %s, seed: %s, size: %f",
								wk.Florist, wk.Soil, wk.Seed, wk.Size)
	}

	fl, err := vouquet.NewFlorist(wk.Florist, pl, status, log)
	if err != nil {
		return nil, err
	}
	fl.SetSize(wk.Size)

	return &Worker{
		fl: fl,
		st_ch: make(chan *farm.State),
		work: wk,

		log: log,
		mtx: lock.NewTryMutex(b_ctx),
		ctx: ctx,
		cancel: cancel,
	}, nil
}

func (self *Worker) GetTarget() (string, string) {
	return self.work.Soil, self.work.Seed
}

func (self *Worker) PostState(state *farm.State) {
	ok, err := self.mtx.TryLock()
	if err != nil {
		if err == lock.ERR_CONTEXT_CANCEL {
			return
		}
		return
	}
	if !ok {
		return
	}

	go func() {
		defer self.mtx.Unlock()

		if self.before_state != nil {
			if self.before_state.Date().Equal(state.Date()) {
				self.log.WriteErr("[%s %s] got same the time in state %s.",
							self.work.Florist, self.work.Soil, self.work.Seed)
				return
			}
		}
		self.before_state = state

		select {
		case <- self.ctx.Done():
		case self.st_ch <- state:
		}
	}()
}

func (self *Worker) Run() error {
	return self.fl.Run(self.ctx, self.st_ch)
}

func (self *Worker) Done() error {
	self.cancel()
	close(self.st_ch)
	return self.fl.Release()
}

func florister() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := new(logger)

	cfg, err := farm.LoadConfig(Cpath)
	if err != nil {
		return err
	}
	r, err := farm.OpenRegistry(cfg, ctx, log)
	if err != nil {
		return err
	}
	defer r.Close()

	log.WriteMsg("Start %s %s", SELF_NAME, Version)

	now := time.Now()
	start := now.AddDate(0, 0, -1)
	workers := []*Worker{}
	for _, work := range cfg.Works {
		init_status, err := r.GetStatus(work.Soil, work.Seed, start, now)
		if err != nil {
			return err
		}

		worker, err := NewWorker(ctx, log, cfg, work, init_status)
		if err != nil {
			return err
		}
		workers = append(workers, worker)

		go func(worker *Worker) {
			if err := worker.Run(); err != nil {
				log.WriteErr("%s", err)
			}
			defer worker.Done()
		}(worker)
	}

	mtx := lock.NewTryMutex(ctx)
	t := time.NewTicker(time.Second)
	for {
		select {
		case <- ctx.Done():
			return nil
		case <-t.C:
			ok, err := mtx.TryLock()
			if err != nil {
				if err == lock.ERR_CONTEXT_CANCEL {
					return nil
				}
				return err
			}
			if !ok {
				continue
			}

			for _, worker := range workers {
				go func(worker *Worker) {
					soil, seed := worker.GetTarget()
					state, err := r.GetLastState(soil, seed)
					if err != nil {
						log.WriteErr("Cannot get status: '%s'", err)
						return
					}

					worker.PostState(state)
				}(worker)
			}

			mtx.Unlock()
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

	if c_path == "" {
		die("empty path")
	}
	Cpath = c_path
}

func main() {
	if err := florister(); err != nil {
		die("%s", err)
	}

}
