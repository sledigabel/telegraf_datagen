package main

import (
	"fmt"
	"math/rand"
	"time"
)

// date format input
const shortForm = "2006-01-02"

type MetricInt struct {
	name string
	tags string
	value int64
}

func (m *MetricInt) String() string{
	return fmt.Sprintf("%s,%s value=%d",m.name,m.tags,m.value)
}

func (m *MetricInt) change(t int64) {
	m.value += rand.Int63n(10)-4
}

func NewMetricInt(name string, tags map[string]string) *MetricInt {
	var str string
	for k,v := range tags {
		if len(str) != 0 {
			str += ","
		}
		str += fmt.Sprintf("%s=%s",k,v)
	}
	return &MetricInt{name,str,rand.Int63n(1000)}
}

type MetricFloat struct {
	name string
	tags string
	value float64
}

func (m *MetricFloat) String() string{
	return fmt.Sprintf("%s,%s value=%.4f",m.name,m.tags,m.value)
}

func (m *MetricFloat) change(t int64) {
	m.value += rand.NormFloat64()*10-4
}

func NewMetricFloat(name string, tags map[string]string) *MetricFloat {
	var str string
	for k,v := range tags {
		if len(str) != 0 {
			str += ","
		}
		str += fmt.Sprintf("%s=%s",k,v)
	}
	return &MetricFloat{name,str,rand.NormFloat64()*100}
}

type MetricBool struct {
	name string
	tags string
	value bool
}

func (m *MetricBool) String() string{
	return fmt.Sprintf("%s,%s value=%t",m.name,m.tags,m.value)
}

func (m *MetricBool) change(t int64) {
	m.value = rand.Intn(20) < 16
}

func NewMetricBool(name string, tags map[string]string) *MetricBool {
	var str string
	for k,v := range tags {
		if len(str) != 0 {
			str += ","
		}
		str += fmt.Sprintf("%s=%s",k,v)
	}
	return &MetricBool{name,str,true}
}


type Metric interface {
	String() string
	change(t int64)
}

type MetricFactory struct {
	metricList []Metric
	timestamp int64
	step int64
}

func (mf *MetricFactory) export(c chan string) {
	for _, metric := range mf.metricList {
		c <- fmt.Sprintf("%s %d",metric.String(),mf.timestamp)
	}
	close(c)
}


var ship chan string
var frequencyMs int64 = 60000
var metrics []Metric

func main(){

	// TODO: add flags and parameters
	var t int64
	var step int64 = frequencyMs * 1000000
	ship = make(chan string,100)
	//var numMetrics int = 100
	//var numTags = 3
	//var numUniqueTags = 3
	//var lenTags=5


	timestp, err := time.Parse(shortForm,"1981-08-04")
	if err != nil {
		panic(err)
	}
	t = timestp.UnixNano()
	tags := make(map[string]string,2)
	tags["tag1"] = "val1"
	tags["tag2"] = "val2"
	//metrics := make([]Metric,2)
	metrics = append(metrics,NewMetricInt("abc.val.int",tags))
	metrics = append(metrics,NewMetricFloat("abc.val.float",tags))

	mf := MetricFactory{metrics,t,step}
	mf.export(ship)

	for s := range ship {
		fmt.Println(s)
	}

}