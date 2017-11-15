package main

import (
	//"fmt"
	"math/rand"
	"time"
	 metrics "./metrics"
	"fmt"
)

// date format input
const (
	longForm = "2006-01-02 15:04:05"
	shortForm = "2006-01-02"
)

func ParseTimeStamp(s string) (time.Time, error) {
	// try with the long form first
	t, err := time.Parse(longForm,s)
	if err != nil {
		return time.Parse(shortForm,s)
	}
	return t,err
}

var ship chan string
var inMs int64 = 1000000
var done chan bool

func readChannel(input chan string){
	//fmt.Println("Start reading")
	for item := range input {
		fmt.Println(item)
	//for range input {
	}
	done <- true
}

func main(){

	// TODO: add flags and parameters
	var (
		t1 int64
		t2 int64
	)
	done = make(chan bool)
	var step int64 = 60000 * inMs
	ship = make(chan string,100)
	rand.Seed(time.Now().Unix())

	timestp, err := ParseTimeStamp("2017-11-12")
	if err != nil {
		panic(err)
	}
	t1 = timestp.UnixNano()

	timestp2, err := ParseTimeStamp("2017-11-13 18:00:00")
	if err != nil {
		panic(err)
	}
	t2 = timestp2.UnixNano()


	mandTags := make(map[string]int,3)
	mandTags["dc"] = 5
	mandTags["hostname"] = 20
	mandTags["env"] = 3

	c := metrics.NewConfigSet()
	c.NumMetrics = 20000
	c.NumTags = 300
	c.MandatoryTags = mandTags
	c.Start = t1
	c.End = t2
	c.Step = step

	//fmt.Println("Creating metrics")
	metricFactory := metrics.NewMetricFactory(c)

	//fmt.Println(metricFactory)
	go readChannel(metricFactory.Output)

	metricFactory.Produce()
	<- done

}