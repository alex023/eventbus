# eventbus
Eventbus is an enhanced version of the standalone pub-sub asynchronous messaging framework.

[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/alex023/eventbus)](https://goreportcard.com/report/github.com/alex023/eventbus)
[![GoDoc](https://godoc.org/github.com/alex023/eventbus?status.svg)](https://godoc.org/github.com/alex023/eventbus)
[![Build Status](https://travis-ci.org/alex023/eventbus.svg?branch=dev)](https://travis-ci.org/alex023/eventbus?branch=dev)
[![Coverage Status](https://coveralls.io/repos/github/alex023/eventbus/badge.svg?branch=dev)](https://coveralls.io/github/alex023/eventbus?branch=dev)

## Brief
消息系统在业务系统中，能够减少对象依赖，避免因为过多聚集、关联带来的维护困难等问题。但传统的[push Msg --> Call Consumers Handle]的同步方式，会有以下个别限制：
- 当消费慢于推送，可能会阻塞等待。
- 消费者无法向包含了自身对象（或方法）的主题推送消息。
- 消费者基于消息更改自身的订阅状态时，可能死锁。
- 在消费者获取消息前，无法进行消息预处理。
- ……

面对这些问题，eventbus 基于actor的异步思想，来处理以上问题，除了最基本的消息推送、订阅功能支持之外，
 也适用于简单的ECS应用场景。具有以下特征：
- 消息的推送与消费隔离，且有缓冲，因此支持：
    - 消息发布更快，不用等待消费者执行
    - 允许消费者向自己推送消息。
    - 推送快于消费时，具备一定的韧性。
- 隔离业务崩溃
- 主题插件支持，以拦截、路由、记录、修改某个主题接受到的消息，并可以：
    - 任意时间添加或卸载（优先于一般消息）
    - [todo]实现消息接受、完成的动态监控

## Can I use it?
 The implementation is still in beta, we are using it for our production already. But the API change will be happen until 1.0.

## Examples
###  base function
本代码演示了三个基本功能：
1. 消息订阅、消费
2. 崩溃隔离
3. 等待消费完成的关闭
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
### topic message filter 
以下代码演示了针对某个具体Topic添加filter的方式，判定被除数是否为零。
除了数据安全验证外，我们还可以通过filter：
1. 记录日志
2. 数据拦截及路由跳转
3. 数据修改（游戏场景中，英雄对某方士兵加血等特效）
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

### [more examples][1]
                       
                       
[1]: https://github.com/alex023/eventbus/tree/dev/example