
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/enginebi/stock"
	"gopkg.in/yaml.v2"
)

type TStock struct {
	Id     string  `yaml:"id"`
	Name   string  `yaml:"name"`
	Target float64 `yaml:"target"`
}

var TStocks []TStock

func main() {
	conf, err := ioutil.ReadFile("/Users/toukii/stock.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(conf, &TStocks)
	if err != nil {
		panic(err)
	}
	// TStocks = append(TStocks, TStock{Id: "111"})
	// o, _ := yaml.Marshal(TStocks)
	// fmt.Println(string(o))

	ids := make([]string, 0, len(TStocks))
	for _, v := range TStocks {
		ids = append(ids, v.Id)
	}

	bs, err := stock.Get(ids...)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	pbs := bytes.Split(bs, []byte("\n"))
	for i, v := range pbs {
		if i >= len(TStocks) {
			break
		}
		ts := TStocks[i]
		_, _, data := stock.Parse(string(v), ",")
		if m := MatchTarget(data, ts); m {
			// fmt.Printf("@%s\n", times)
			// alert("")
		}
	}
}

func MatchTarget(data []float64, ts TStock) bool {
	avg := stock.Favg(data)
	avg2 := stock.Fposavg(data)
	ravg2 := stock.Fnegavg(data)
	norm := stock.Fnorm(data, avg)

	fmt.Printf("### %s(%s)\n\t target: %.2f, avg: %.2f(%.2f ~ %.2f) norm: %.2f\n", ts.Id, ts.Name, ts.Target, avg, ravg2, avg2, norm)
	if avg >= ts.Target*1.03 {
		return false
	}
	return true
}