package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"time"
)

type CountMessage struct {
	X, Y int
}
type Pubslisher struct {
}
type Consumer struct {
	testNilMap map[int32]struct{}
}

func (c *Consumer) Add(message interface{}) {
	cMsg, ok := message.(CountMessage)
	if !ok {
		return
	}
	if cMsg.X == 95 {
		c.testNilMap[2] = struct{}{}
	} else {
		fmt.Printf("result:=%v + %v = %v \n", cMsg.X, cMsg.Y, cMsg.X+cMsg.Y)

	}
}

func main() {
	eb := eventbus.New()
	consumer := &Consumer{}
	eb.Subscribe("add", "consumer", consumer.Add)
	time.Sleep(time.Second)
	fmt.Println("topics:", eb.GetTopics())
	for i := 0; i < 100; i++ {
		eb.Publish("add", CountMessage{i, i * 2})
	}
	time.Sleep(time.Second)

	for i := 100; i < 200; i++ {
		eb.Publish("add", CountMessage{i, i * 2})
	}
}
