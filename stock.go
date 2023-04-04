
package stock

import (
	// "bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	// "github.com/chfenger/goNum"
)

var (
	uri = "http://hq.sinajs.cn/list="
)

func Fnorm(data []float64, avg float64) float64 {
	size := len(data)
	if size <= 0 {
		return 0
	}
	sum := float64(0)
	for _, v := range data {
		sum += math.Pow(v-avg, 2)
	}
	a := sum / float64(size)
	return math.Sqrt(a)
}

func Fposavg(data []float64) float64 {
	size := len(data)
	if size <= 0 {
		return 0
	}
	avg := data[0]
	for i := 1; i < size; i++ {
		avg = (avg + data[i]) / 2.0
	}
	return avg
}

func Fnegavg(data []float64) float64 {
	size := len(data)
	if size <= 0 {
		return 0
	}
	avg := data[size-1]
	for i := size - 2; i >= 0; i-- {
		avg = (avg + data[i]) / 2.0
	}
	return avg
}

func Favg(data []float64) float64 {
	size := len(data)
	if size <= 0 {
		return 0
	}
	sum := float64(0)
	for _, v := range data {
		sum += v
	}
	return sum / float64(size)
}

func Parse(hq string, delimiter string) (title string, times string, data []float64) {
	phq := strings.Split(hq, delimiter)
	if len(phq) <= 1 {
		fmt.Println(hq)
		return
	}
	var vs []string
	vs = phq
	if strings.Contains(phq[0], "=") {
		title = strings.Split(phq[0], "=\"")[1]
		start := 1
		end := len(phq) - 4
		times = strings.Join(phq[end:], " ")
		vs = phq[start:end]
	}

	data = make([]float64, 0, len(vs))
	m := make(map[int]int, len(vs))
	offset := 80
	for _, it := range vs {
		it = strings.Trim(it, "\n")
		v, err := strconv.ParseFloat(it, 64)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if v <= 0 {
			continue
		}
		m[int(v)/offset*offset]++
		data = append(data, v)
	}
	mv := 0
	mk := 0
	for k, v := range m {
		if v > mv {
			mv = v
			mk = k
		}
	}
	match := func(v float64) bool {
		if iv := int(v); mk-offset <= iv && iv <= mk+offset {
			return true
		}
		return false
	}
	idx := 0
	for i, v := range data {
		if !match(v) {
			continue
		}
		if idx != i {
			data[idx] = v
		}
		idx++
	}
	data = data[:idx]
	fmt.Println(data)
	return
}

func Get(stockids ...string) ([]byte, error) {
	furi := uri + strings.Join(stockids, ",")
	resp, err := http.Get(furi)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}