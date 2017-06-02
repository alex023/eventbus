package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"math/rand"
	"sync/atomic"
	"time"
)

type CountMessage struct {
	X, Y int
}

//DivisorJudgment implement struct of "Filter"
type DivisorJudgment struct {
}

func (m *DivisorJudgment) HandleMessage(message interface{}) (proceed bool) {
	msg, ok := message.(CountMessage)
	if !ok {
		return false
	}
	//determines whether the divisor is zero。if it's ,catch it.
	if msg.Y == 0 {
		fmt.Printf("[Filter] [X/Y]:[%2d/%2d]= ,zero the dividend \n", msg.X, msg.Y)
		return false
	}
	return true
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

type Watcher struct {
	name            string
	subscribeNum    int64
	messageReceived uint64
	messagePushed   uint64
	filterNum       int64
}

func (ds *Watcher) SetTopicName(topicName string) {
	ds.name = topicName
}
func (ds *Watcher) TopicSubscribe() {
	atomic.AddInt64(&ds.subscribeNum, 1)
}
func (ds *Watcher) TopicUnscribe() {
	atomic.AddInt64(&ds.subscribeNum, -1)
}
func (ds *Watcher) FilterLoad() {
	atomic.AddInt64(&ds.filterNum, 1)
}
func (ds *Watcher) FilterUnload() {
	atomic.AddInt64(&ds.filterNum, -1)
}
func (ds *Watcher) MessagePushed(message interface{}) {
	atomic.AddUint64(&ds.messagePushed, 1)
}

func (ds *Watcher) MessageReceived(message interface{}) {
	atomic.AddUint64(&ds.messageReceived, 1)
}

//此代码演插件对消息的处理，如何影响消息！
func main() {
	var (
		r        = rand.New(rand.NewSource(time.Now().UnixNano()))
		eb       = eventbus.New()
		consumer = &Consumer{}
		topic    = "T"
	)
	eb.InitTopic(topic, &Watcher{})
	eb.Subscribe(consumer.Div, topic)
	eb.LoadFilter(topic, &DivisorJudgment{})

	fmt.Println("catch zero divisor by filter.")

	//0...49,catch zero divisor by [watcher];and 50...99,no watcher.
	for i := 0; i < 100; i++ {
		eb.Push(topic, CountMessage{i, r.Intn(5)})
		time.Sleep(time.Millisecond * 50)
		if i == 50 {
			fmt.Printf("[topic statics]:%+v \n", eb.Statistic(topic)[0])

			fmt.Println("recover collapse by eventbus.")
			eb.UnloadFilter(topic, &DivisorJudgment{})
		}
	}
	eb.StopGracefull()
	fmt.Printf("[topic statics]:%+v \n", eb.Statistic(topic)[0])

}
