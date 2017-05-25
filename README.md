# eventbus
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/alex023/eventbus)](https://goreportcard.com/report/github.com/alex023/eventbus)
[![GoDoc](https://godoc.org/github.com/alex023/eventbus?status.svg)](https://godoc.org/github.com/alex023/eventbus)
[![Build Status](https://travis-ci.org/alex023/eventbus.svg?branch=dev)](https://travis-ci.org/alex023/eventbus?branch=dev)
[![Coverage Status](https://coveralls.io/repos/github/alex023/eventbus/badge.svg?branch=dev)](https://coveralls.io/github/alex023/eventbus?branch=dev)


Eventbus is an enhanced version of the standalone pub-sub asynchronous messaging framework.

eventbus 是一个加强版的单机pub-sub异步消息框架。除了一般的消息订阅、发布的功能支持之外，
 也适用于简单的ECS应用场景。具有以下特征：
- 大吞吐量，不低于200wQPS
- 隔离业务崩溃
- 主题[过滤器]支持,且可以在任意时间添加或卸载
- [todo]考虑提供统计插件支持

基于该框架，可以将对象间的直接依赖改为消息通知，而避免因为过多聚集、关联带来的维护困难、对象释放不彻底等问题。

## Example: base function
```golang
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
```
## Example：filter
```golang

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

```

>more examples,can see dir example