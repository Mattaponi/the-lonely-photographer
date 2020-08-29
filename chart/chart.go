
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/enginebi/stock"
	"github.com/qiniu/iconv"
	"github.com/zserge/lorca"
)

var (
	buf9  [512]byte
	buf10 [1024]byte
	buf11 [2048]byte
	buf12 [4096]byte

	curved    chan *CurveD
	delimiter string
)

func init() {
	curved = make(chan *CurveD, 3)

	flag.StringVar(&delimiter, "d", ",", "-d ,  \nnote: delimiter 分隔符")
}

func app() {
	ui, err := lorca.New("", "", 350, 240)
	if err != nil {
		panic(err)
	}

	ui.Bind("curve", Curve)
	defer ui.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	go http.Serve(ln, http.FileServer(stock.FS))
	// go http.Serve(ln, http.FileServer(http.Dir("./")))
	ui.Load(fmt.Sprintf("http://%s", ln.Addr()))

	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}
}

func main() {
	flag.Parse()
	//先取程序的标准输入属性信息
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	// 判断标准输入设备属性 os.ModeCharDevice 是否设置
	// 同时判断是否有数据输入
	if (info.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe {
		bs, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		ubs := charset(bs)
		title, _, data := stock.Parse(string(ubs), delimiter)
		fmt.Fprintf(os.Stdout, "title: %s, data: %+v\n", title, data)

		avg := stock.Favg(data)
		pavg, navg := stock.Fposavg(data), stock.Fnegavg(data)
		curved <- &CurveD{
			Title:  title,
			Data:   data,
			Avg:    avg,
			PosAvg: pavg,
			NegAvg: navg,
		}
	}
	app()
}

func charset(bs []byte) []byte {
	icv, err := iconv.Open("utf-8", "gbk")
	if err != nil {
		panic(err)
	}
	defer icv.Close()

	var obuf []byte
	lbs := len(bs)
	if lbs <= 500 {
		obuf = buf9[:]
	} else if lbs <= 1000 {
		obuf = buf10[:]
	} else if lbs <= 2000 {
		obuf = buf11[:]
	} else if lbs <= 4096 {
		obuf = buf12[:]
	} else {
		obuf = make([]byte, lbs+512)
	}

	out, _, err := icv.Conv(bs, obuf[:])
	if err != nil {
		panic(err)
	}
	return out
}

// curl http://hq.sinajs.cn/list\=sh600745 | chart
type CurveD struct {
	Title  string    `json:"title"`
	Data   []float64 `json:"data"`
	Avg    float64   `json:"avg"`
	PosAvg float64   `json:"posavg"`
	NegAvg float64   `json:"negavg"`
}

func Curve() string {
	cd := <-curved
	defer func() {
		curved <- cd
	}()
	bs, err := json.Marshal(cd)
	if err != nil {
		return err.Error()
	}
	fmt.Fprintf(os.Stdout, "%s\n", bs)
	return string(bs)
}