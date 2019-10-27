package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

var channel = "myChannel"

func write(str string) {
	f, err := os.OpenFile("/GitHub/concurrent-file-processing/data/sample.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(f, "", 0)
	logger.Output(2, str)
}

// listenPubSubChannels listens for messages on Redis pubsub channels.
// The onMessage function is called for each message.
func listenPubSubChannels(ctx context.Context, redisServerAddr string,
	onMessage func(channel string, data []byte) error,
	channels string) error {

	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	const healthCheckPeriod = time.Minute

	c, err := redis.Dial("tcp", redisServerAddr, // Read timeout on server should be greater than ping period.
		redis.DialReadTimeout(healthCheckPeriod+10*time.Second),
		redis.DialWriteTimeout(10*time.Second))

	if err != nil {
		return err
	}

	defer c.Close()

	psc := redis.PubSubConn{Conn: c}

	if err := psc.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
		return err
	}

	done := make(chan error, 1)

	// Start a goroutine to receive notifications from the server.
	go func() {
		for {
			switch n := psc.Receive().(type) {
			case error:
				done <- n
				return
			case redis.Message:
				if err := onMessage(n.Channel, n.Data); err != nil {
					done <- err
					return
				}
			}
		}
	}()

	ticker := time.NewTicker(healthCheckPeriod)
	defer ticker.Stop()
loop:
	for err == nil {
		select {
		case <-ticker.C:
			// Send ping to test health of connection and server.
			// If corresponding pong is not received, then receive on the
			// connection will timeout and the receive goroutine will exit.
			if err = psc.Ping(""); err != nil {
				break loop
			}
		case <-ctx.Done():
			break loop
		case err := <-done:
			// Return error from the receive goroutine.
			return err
		}
	}

	// Signal the receiving goroutine to exit by unsubscribing from all channels.
	psc.Unsubscribe()

	// Wait for goroutine to complete.
	return <-done
}

// This example shows how receive pubsub notifications with cancelation and
// health checks.
func main() {
	ctx, cancel := context.WithCancel(context.Background())

	err := listenPubSubChannels(ctx,
		"localhost:6379",
		func(channel string, message []byte) error {
			fmt.Printf("channel: %s, message: %s\n", channel, message)

			// For the purpose of this example, cancel the listener's context
			// after receiving last message sent by publish().
			if string(message) == "End" {
				cancel()
			} else {
				write(string(message))
			}
			return nil
		}, channel)

	if err != nil {
		fmt.Println(err)
		return
	}
}
