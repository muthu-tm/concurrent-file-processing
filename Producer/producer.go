package main

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

var channel = "myChannel"
var PUBLISH = "PUBLISH"

func publish() {
	c, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	c.Do(PUBLISH, channel, "Line 1")
	c.Do(PUBLISH, channel, "Line 2")
	c.Do(PUBLISH, channel, "Line 3")
	c.Do(PUBLISH, channel, "End")
}

func main() {
	publish()
}
