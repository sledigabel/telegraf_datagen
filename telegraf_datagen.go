package main

import (
	//"fmt"
	"math/rand"
	"time"
	 metrics "./metrics"
)

// date format input
const shortForm = "2006-01-02 03:04:05"


var ship chan string
var frequencyMs int64 = 60000
var done chan bool

func readChannel(input chan string){
	//fmt.Println("Start reading")
	//for item := range input {
		//fmt.Println(item)
	for range input {
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
	var step int64 = frequencyMs * 1000000
	ship = make(chan string,100)
	//var numMetrics int = 100
	//var numTags = 3
	//var numUniqueTags = 3
	//var lenTags=5
	rand.Seed(time.Now().Unix())

	timestp, err := time.Parse(shortForm,"1981-08-04 00:00:00")
	if err != nil {
		panic(err)
	}
	t1 = timestp.UnixNano()

	timestp2, err := time.Parse(shortForm,"1981-08-04 0:00:00")
	if err != nil {
		panic(err)
	}
	t2 = timestp2.UnixNano()


	mandTags := make(map[string]int,3)
	mandTags["dc"] = 5
	mandTags["hostname"] = 300
	mandTags["env"] = 3

	//fmt.Println("Creating metrics")
	metricFactory := metrics.NewMetricFactory(400000,150, mandTags,t1,step,t2)

	//fmt.Println(metricFactory)
	go readChannel(metricFactory.Output)

	metricFactory.Produce()
	<- done

}