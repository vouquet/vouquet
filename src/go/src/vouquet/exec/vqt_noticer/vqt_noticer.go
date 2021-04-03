package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"context"
)

import (
	"vouquet/farm"
	"vouquet/advertizer"
)

const (
	SELF_NAME string = "vqt_noticer"
	USAGE string = "[-version] [-c <config path>] <Path of Credentical Twitter> <SEED> <SOIL>"

	TW_WAIT_SEC int64 = 5
)

var (
	Version string
	Cpath   string
	TwCpath string
	Seed  string
	Soil    string
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

func noticer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := new(logger)

	twc, err := advertizer.NewTwitterClient(TwCpath)
	if err != nil {
		return err
	}

	s_recorder, err := farm.OpenShipRecorder(Soil, Seed, Cpath, ctx, log)
	if err != nil {
		return err
	}
	sr_ch, err := s_recorder.StreamRead()
	if err != nil {
		return err
	}

	wakeup_tstr := time.Now().Format("2006/01/02 15:04")
	msg := fmt.Sprintf("稼働状態通知を開始しました。\n起動時刻は、%sです\n取引累積値が初期化されました。\n¥0から再カウントします。", wakeup_tstr)
	if err := twc.Tweet(msg); err != nil {
		return err
	}

	log.WriteMsg("Start %s %s", SELF_NAME, Version)
	var date_str string = time.Now().Format("2006/01/02")
	var date_yield float64
	var total_yield float64
	for {
		select {
		case <- ctx.Done():
			return nil
		case sr, ok := <- sr_ch:
			if !ok {
				return nil
			}
			if sr == nil {
				return nil
			}

			now_str := time.Now().Format("2006/01/02")
			if date_str != now_str {
				date_str = now_str
				date_yield = 0
			}

			var msg string
			if sr.IsOpenOrder() {
				var jp_o_str string
				var val_type string
				if sr.OrderType() == farm.TYPE_SELL {
					jp_o_str = "ショート"
					val_type = "bid"
				}
				if sr.OrderType() == farm.TYPE_BUY {
					jp_o_str = "ロング"
					val_type = "ask"
				}

				msg = fmt.Sprintf("%s ポジション 作成\nエントリー価格(%s): %.3f",
										jp_o_str, val_type, sr.Price())
			} else {
				var jp_o_str string
				var val_type string
				if sr.OrderType() == farm.TYPE_BUY {
					jp_o_str = "ショート"
					val_type = "ask"
				}
				if sr.OrderType() == farm.TYPE_SELL {
					jp_o_str = "ロング"
					val_type = "bid"
				}

				date_yield += sr.Yield()
				total_yield += sr.Yield()
				msg = fmt.Sprintf("%s ポジション 利確\n利確価格(%s): %.3f\n\n今回の利益: ¥%.1f\n%sの利益合計: ¥%.1f\n起動(%s)からの総額: ¥%.1f",
					jp_o_str, val_type, sr.Price(), sr.Yield(),
					date_str, date_yield,
					wakeup_tstr, total_yield,)
			}

			if err := twc.Tweet(msg); err != nil {
				log.WriteErr("Failed tweet: '%s'", err)
			}
			time.Sleep(time.Second * time.Duration(TW_WAIT_SEC))
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

	if flag.NArg() < 3 {
		die("usage : %s %s", SELF_NAME, USAGE)
	}

	tw_cpath := flag.Arg(0)
	seed := flag.Arg(1)
	soil := flag.Arg(2)

	if tw_cpath == "" {
		die("empty path of credential twitter")
	}
	if c_path == "" {
		die("empty path")
	}

	Cpath = c_path
	TwCpath = tw_cpath
	Seed = seed
	Soil = soil
}

func main() {
	if err := noticer(); err != nil {
		die("%s", err)
	}
}
