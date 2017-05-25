package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"time"
)

type CountMessage struct {
	X, Y int
}

var (
	eb       = eventbus.New()
	topic    = "add"
	clientId = "consumer"
)

type Consumer struct {
	testNilMap map[int32]struct{}
	counter    int
}

func (c *Consumer) Add(message interface{}) {
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
		eb.Unsubscribe(topic, clientId)
	}
}

//此代码演示消费者订阅的基本功能：
// 1.故障隔离
// 2.内部取消订阅，这可以用于特殊情况下的状态切换
func main() {
	consumer := &Consumer{}
	eb.Subscribe(topic, clientId, consumer.Add)
	time.Sleep(time.Second)
	fmt.Println("topics:", eb.GetTopics())
	for i := 0; i < 100; i++ {
		eb.Publish("add", CountMessage{i, i * 2})
	}
	time.Sleep(time.Second)

	//those messages cannot received by consumer,because it unsubscribe this topic
	for i := 100; i < 200; i++ {
		eb.Publish("add", CountMessage{i, i * 2})
	}

	eb.StopGracefull()
}
