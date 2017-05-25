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
}

func (c *Consumer) Add(message interface{}) {
	cMsg, ok := message.(CountMessage)
	if !ok {
		return
	}
	//当输入值 %5==0时，故意向nil map 插入值导致崩溃
	if cMsg.X%5 == 0 {
		c.testNilMap[2] = struct{}{}
	} else {
		fmt.Printf("result:=%v + %v = %v \n", cMsg.X, cMsg.Y, cMsg.X+cMsg.Y)

	}
}

//此代码演示消费者订阅，且在某特殊情况下崩溃，系统对其的处理！
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
