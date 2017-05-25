package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"time"
)

type CountMessage struct {
	X, Y int
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
	//当输入值 %10==0时，故意向nil map 插入值导致崩溃
	if cMsg.X%10 == 0 {
		c.testNilMap[2] = struct{}{}
	} else {
		fmt.Printf("result:=%v + %v = %v \n", cMsg.X, cMsg.Y, cMsg.X+cMsg.Y)

	}

	if c.counter == 99 {
		c.sub.Unscribe()
	}
}

//此代码演示消费者订阅的基本功能：
// 1.故障隔离
// 2.内部取消订阅，这可以用于特殊情况下的状态切换
func main() {
	var (
		eb    = eventbus.New()
		topic = "add"
	)
	consumer := &Consumer{}
	sub, _ := eb.Subscribe(topic, consumer.HandleMessage)
	consumer.sub = sub
	time.Sleep(time.Second)
	for i := 0; i < 100; i++ {
		eb.Publish(topic, CountMessage{i, i * 2})
	}
	time.Sleep(time.Second)

	//those messages cannot received by consumer,because it unsubscribe this topic
	for i := 100; i < 200; i++ {
		eb.Publish(topic, CountMessage{i, i * 2})
	}

	eb.StopGracefull()
}
