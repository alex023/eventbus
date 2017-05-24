package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"time"
)

type Consumer struct {
	counter int
}

const _COUNT_SIZE = 50000000

func (c *Consumer) Count(msg interface{}) {
	c.counter++
	if c.counter == _COUNT_SIZE {
		c.Info()
	}
}
func (c *Consumer) Info() {
	fmt.Printf("receive msg :%4.4f \n", float32(c.counter)/10000)
}
func main() {
	eb := eventbus.New()
	consumer := &Consumer{}
	eb.Subscribe("counter", "1", consumer.Count)
	fmt.Println("sending 5000w message begin.......")
	start := time.Now()
	for i := 0; i < _COUNT_SIZE; i++ {
		eb.Publish("counter", struct{}{})
	}
	consumer.Info()
	time.Sleep(time.Second)
	end := time.Now()

	fmt.Printf("spent：%2.0f seconds \n", end.Sub(start).Seconds())
}
