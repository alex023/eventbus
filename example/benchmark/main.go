package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"time"
)

type Consumer struct {
	counter int
}

func (c *Consumer) Count(msg interface{}) {
	c.counter++
}
func (c *Consumer) Info(msg interface{}) {
	fmt.Println("receive msg :", c.counter/10000, "w")
}
func main() {
	eb := eventbus.New()
	consumer := &Consumer{}
	eb.Subscribe("counter", "1", consumer.Count)
	fmt.Println("sending 5000w message begin.......")
	start := time.Now()
	for i := 0; i < 50000000; i++ {
		eb.Publish("counter", struct{}{})
	}
	end := time.Now()
	consumer.Info(struct{}{})
	fmt.Printf("spent：%2.0f seconds \n", end.Sub(start).Seconds())
}
