package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"time"
)

type CountMessage struct {
	Num int
}

type Consumer struct {
	testNilMap map[int32]struct{}
	counter    int
	sub        *eventbus.Subscribe
}

func (c *Consumer) HandleMessage(message interface{}) {
	cMsg, ok := message.(CountMessage)
	if !ok {
		return
	}
	c.counter++
	//When a condition is met, performe the operation which causes the crash
	if cMsg.Num == 6 {
		c.testNilMap[2] = struct{}{}
	} else {
		fmt.Printf("Num:=%v  \n", cMsg.Num)

	}

	if c.counter > 9 {
		fmt.Println("consumer unscribe topic,and cannot receive any message later.")
		c.sub.Unscribe()
	}
}

//此代码演示消费者订阅的基本功能：
// 1.fault isolation
// 2.the subscription object itself is unsubscribed
// 3.eventbus stop gracefull
func main() {
	var (
		eb    = eventbus.New()
		topic = "add"
	)
	consumer := &Consumer{}
	sub, _ := eb.Subscribe(topic, consumer.HandleMessage)
	consumer.sub = sub

	fmt.Println("send 10 messages, Num from 0 to 9")
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 500)
		eb.Publish(topic, CountMessage{i})
	}

	time.Sleep(time.Second)
	fmt.Println("send 10 messages, Num from 10 to 19")
	//those messages cannot received by consumer,because it unsubscribe this topic
	for i := 10; i < 20; i++ {
		time.Sleep(time.Millisecond * 200)
		eb.Publish(topic, CountMessage{i})
	}

	eb.StopGracefull()
}
