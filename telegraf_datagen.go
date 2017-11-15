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
	inMs int64 = 1000000
)

var ship chan string
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

	done = make(chan bool)
	var step int64 = 60000 * inMs
	ship = make(chan string,100)
	rand.Seed(time.Now().Unix())

	timestp := "2017-11-12"
	timestp2 := "2017-11-13 18:00:00"

	mandTags := make(map[string]int,3)
	mandTags["dc"] = 5
	mandTags["hostname"] = 20
	mandTags["env"] = 3

	c := metrics.NewConfigSet()
	c.NumMetrics = 20000
	c.NumTags = 300
	c.MandatoryTags = mandTags
	c.Start = timestp
	c.End = timestp2
	c.Step = step

	//fmt.Println("Creating metrics")
	metricFactory := metrics.NewMetricFactory(c)

	//fmt.Println(metricFactory)
	go readChannel(metricFactory.Output)

	metricFactory.Produce()
	<- done

}