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
	"vouquet/advertizer"
)

var (
	Cpath   string
	TwCpath string
	Symbol  string
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

	s_recorder, err := soil.OpenShipRecorder(Soil, Symbol, Cpath, ctx, log)
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

			go func(sr *soil.ShipRecord) {
				var msg string
				if sr.IsOpenOrder() {
					var jp_o_str string
					var val_type string
					if sr.OrderType() == soil.TYPE_SELL {
						jp_o_str = "ショート"
						val_type = "bid"
					}
					if sr.OrderType() == soil.TYPE_BUY {
						jp_o_str = "ロング"
						val_type = "ask"
					}

					msg = fmt.Sprintf("%s ポジション 作成, エントリー価格(%s): %.3f",
											jp_o_str, val_type, sr.Price())
				} else {
					var jp_o_str string
					var val_type string
					if sr.OrderType() == soil.TYPE_BUY {
						jp_o_str = "ショート"
						val_type = "ask"
					}
					if sr.OrderType() == soil.TYPE_SELL {
						jp_o_str = "ロング"
						val_type = "bid"
					}

					total_yield += sr.Yield()
					msg = fmt.Sprintf("%s ポジション 利確\n利確価格(%s): %.3f\n\n今回の利益: %.3f\n%sからの利益総額: %.3f",
						jp_o_str, val_type, sr.Price(), sr.Yield(), wakeup_tstr, total_yield)
				}

				if err := twc.Tweet(msg); err != nil {
					log.WriteErr("Failed tweet: '%s'", err)
				}
			}(sr)
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
	flag.StringVar(&c_path, "c", "./vouquet.conf", "config path.")
	flag.Parse()

	if flag.NArg() < 3 {
		die("usage : vqt_florister [-c <config path>] <Path of Credentical Twitter> <SYMBOL> <SOIL>")
	}

	tw_cpath := flag.Arg(0)
	symbol := flag.Arg(1)
	soil := flag.Arg(2)

	if tw_cpath == "" {
		die("empty path of credential twitter")
	}
	if c_path == "" {
		die("empty path")
	}

	Cpath = c_path
	TwCpath = tw_cpath
	Symbol = symbol
	Soil = soil
}

func main() {
	if err := noticer(); err != nil {
		die("%s", err)
	}
}
