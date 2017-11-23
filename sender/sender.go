package sender

import (
	"fmt"
	"net"
	"os"
	"time"
)

const (
	Port                = 8094
	MaxRetries          = 3
	BufferSize          = 1000
	SendBufferSizeBytes = 1000
	FlushPeriod         = 1 * time.Second

	// connection states
	STOPPED    = 0
	CONNECTING = 1
	CONNECTED  = 2
)

type Endpoint struct {
	Server              string        `toml:"server"`
	Port                int           `toml:"port"`
	MaxRetries          int           `toml:"connection_max_retries"`
	BufferSize          int           `toml:"sender_buffer_size"`
	SendBufferSizeBytes int           `toml:"connection_buffer_size"`
	FlushPeriod         time.Duration `toml:"connection_flush_period"`
	Send                chan string
	retryCounter        int
	BytesSent           int64
	conn                net.Conn
	Stop                chan bool // any signal would cause a stop.
	State               int
	Debug               bool `toml:"debug"`
}

func NewEndpoint() *Endpoint {
	ep := &Endpoint{}
	ep.Server = "localhost"
	ep.Port = Port
	ep.MaxRetries = MaxRetries
	ep.SendBufferSizeBytes = SendBufferSizeBytes
	// TODO: at the moment you can't override BufferSize in the config file because the make is done here.
	ep.Send = make(chan string, BufferSize)
	ep.FlushPeriod = FlushPeriod
	ep.Stop = make(chan bool, 1)
	ep.State = STOPPED
	ep.Debug = false
	return ep
}

func (e *Endpoint) Connect() error {
	var err error
	e.State = CONNECTING
	if e.retryCounter >= e.MaxRetries {
		fmt.Errorf("maximum connection attempts reached: %d", e.retryCounter)
		os.Exit(127)
	}
	fmt.Println("Connecting to", e.Server, "on", e.Port)
	e.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", e.Server, e.Port))
	if err != nil {
		e.retryCounter++
		// TODO: add some timer here.
		e.Connect()
	} else {
		e.conn.(*net.TCPConn).SetKeepAlive(true)
		// we're connected.
	}
	e.State = CONNECTED
	return nil
}

func (e *Endpoint) SendBytes(b []byte) {
	for _, err := e.conn.Write(b); err != nil; e.retryCounter++ {
		fmt.Errorf("Failed to send buffer. Reconnecting\n")
		e.Connect()
	}
}

func (e *Endpoint) Expedite() {

	flushTick := time.NewTicker(FlushPeriod)
	buffer := make([]byte, 0, e.SendBufferSizeBytes)

	for {
		select {
		//receives a signal
		case <-e.Stop:
			// sends current buffer
			e.SendBytes(buffer)
			e.conn.Close()
			e.State = STOPPED
			return
		// time to flush
		case <-flushTick.C:
			e.SendBytes(buffer)
			buffer = make([]byte, 0, e.SendBufferSizeBytes)

		case item := <-e.Send:

			buffer = append(buffer, []byte(fmt.Sprintf("%s\n", item))...)
			// TODO: better handle the capacity and avoid over allocating
			if len(buffer) >= e.SendBufferSizeBytes {
				// send until successful (will panic if exceeds too many retries
				e.SendBytes(buffer)
				// stats
				e.BytesSent += int64(len(buffer))
				buffer = make([]byte, 0, e.SendBufferSizeBytes)
			}

		}
	}
}
