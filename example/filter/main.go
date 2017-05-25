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

//DivisorJudgment implement struct of "Filter"
type DivisorJudgment struct {
}

func (m *DivisorJudgment) Receive(message interface{}) (newMsg interface{}, proceed bool) {
	msg, ok := message.(CountMessage)
	if !ok {
		return nil, false
	}
	//determines whether the divisor is zero。if it's ,catch it.
	if msg.Y == 0 {
		fmt.Printf("[Filter] [X/Y]:[%2d/%2d]= ,zero the dividend \n", msg.X, msg.Y)
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
		//fmt.Printf("[Consumer] [X/Y]:[%2d/%2d]= %d \n", cMsg.X, cMsg.Y, cMsg.X/cMsg.Y)
		_ = cMsg.X / cMsg.Y

	}
}

//此代码演插件对消息的处理，如何影响消息！
func main() {
	var (
		r        = rand.New(rand.NewSource(time.Now().UnixNano()))
		eb       = eventbus.New()
		consumer = &Consumer{}
	)

	eb.Subscribe("add", consumer.Div)
	eb.LoadFilter("add", &DivisorJudgment{})

	fmt.Println("catch zero divisor by filter.")

	//0...49,catch zero divisor by [filter];and 50...99,no filter.
	for i := 0; i < 100; i++ {
		eb.Publish("add", CountMessage{i, r.Intn(5)})
		time.Sleep(time.Millisecond * 100)
		if i == 50 {
			fmt.Println("recover collapse by eventbus.")
			eb.UnloadFilter("add", &DivisorJudgment{})

		}
	}
	eb.StopGracefull()
}
