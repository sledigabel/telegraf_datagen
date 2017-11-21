package main

import (
	"./metrics"
	"./sender"
	"fmt"
	"github.com/BurntSushi/toml"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// date format input
const (
	inMs           int64  = 1000000
	configFileName string = "telegraf_datagen.conf"
)

var ship chan string
var done chan bool

type DataGen struct {
	mf           *metrics.MetricFactory
	be           *sender.Endpoint
	prevProduced int64
	prevSent     int64
}

func main() {

	// TODO: add flags and parameters

	sigs := make(chan os.Signal, 1)
	done = make(chan bool)
	var step int64 = 60000 * inMs
	ship = make(chan string, 100)
	rand.Seed(time.Now().Unix())

	statPeriod := time.Minute

	timestp := "2017-11-20 17:41:00"
	timestp2 := "2017-11-20 18:00:00"

	c := metrics.NewConfigSet()
	c.NumMetrics = 20000
	c.NumTags = 300
	//c.MandatoryTags = mandTags
	c.Start = timestp
	c.End = timestp2
	c.Step = step

	fmt.Println(c)

	dataGen := &DataGen{}
	dataGen.be = sender.NewEndpoint()

	// TODO: refactor
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		//fmt.Println("No config file")
	} else {
		//fmt.Println("config file found")
		if _, err := toml.DecodeFile(configFileName, &c); err != nil {
			panic(err)
		}
		if _, err := toml.DecodeFile(configFileName, &dataGen.be); err != nil {
			panic(err)
		}
	}

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	dataGen.mf = metrics.NewMetricFactory(c)

	//backend.Send = metricFactory.Output
	dataGen.be.Connect()
	go dataGen.be.Expedite()

	stats := time.NewTicker(statPeriod)

	go dataGen.mf.Produce()
	// transmit the data from the metricFactory to the sender
	for {
		select {
		case <-sigs:
			fmt.Println("Received TERM signal.")
			dataGen.mf.Stop <- true
			dataGen.be.Stop <- true
			// waiting for the connection to flush and stop
			for dataGen.be.State != sender.STOPPED {
			}
			fmt.Println("Exiting")
			os.Exit(1)

		case <-stats.C:
			// captures the instant values
			produced, sent := dataGen.mf.Counter, dataGen.be.BytesSent
			str := fmt.Sprintf("Produced: %d (%.2fK/s), Sent: %.2fKB/s, Queue: %d, timestamp: %s",
				produced,
				float64(produced-dataGen.prevProduced)/statPeriod.Seconds()/1000,
				float64(sent-dataGen.prevSent)/statPeriod.Seconds()/1024,
				len(dataGen.mf.Output),
				dataGen.mf.CurrentTime().Format(time.Stamp))
			fmt.Println(str)
			dataGen.prevProduced, dataGen.prevSent = produced, sent
		case i := <-dataGen.mf.Output:
			dataGen.be.Send <- i
		}
	}
}
