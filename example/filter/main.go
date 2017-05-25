package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"math/rand"
	"time"
)

type CountMessage struct {
	X, Y int
}

//判断除数为0的插件
type DivisorJudgment struct {
}

//当除数为0时，返回false
func (m *DivisorJudgment) Receive(message interface{}) (newMsg interface{}, proceed bool) {
	msg, ok := message.(CountMessage)
	if !ok {
		return nil, false
	}
	//判断除数为0
	if msg.Y == 0 {
		fmt.Printf("[Plugin] [X/Y]:[%2d/%2d]= ,zero the dividend \n", msg.X, msg.Y)
		return nil, false
	}
	return msg, true
}

type Consumer struct {
	testNilMap map[int32]struct{}
}

func (c *Consumer) Div(message interface{}) {
	cMsg, ok := message.(CountMessage)
	if !ok {
		return
	}
	if cMsg.X == 95 {
		c.testNilMap[2] = struct{}{}
	} else {
		fmt.Printf("[Consumer] [X/Y]:[%2d/%2d]= %d \n", cMsg.X, cMsg.Y, cMsg.X/cMsg.Y)

	}
}

//此代码演插件对消息的处理，如何影响消息！
func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	eb := eventbus.New()
	consumer := &Consumer{}
	eb.Subscribe("add", consumer.Div)
	//加载针对某个主题的插件
	eb.LoadFilter("add", &DivisorJudgment{})
	time.Sleep(time.Second)
	fmt.Println("topics:", eb.GetTopics())
	for i := 0; i < 100; i++ {
		y := r.Intn(5)
		eb.Publish("add", CountMessage{i, y})
		//卸载插件
		if i == 50 {
			eb.UnloadFilter("add", &DivisorJudgment{})

		}
	}
	eb.StopGracefull()
}
