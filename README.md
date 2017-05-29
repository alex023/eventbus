# eventbus
[**中文介绍**](https://github.com/alex023/eventbus/wiki)  

[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/alex023/eventbus)](https://goreportcard.com/report/github.com/alex023/eventbus)
[![GoDoc](https://godoc.org/github.com/alex023/eventbus?status.svg)](https://godoc.org/github.com/alex023/eventbus)
[![Build Status](https://travis-ci.org/alex023/eventbus.svg?branch=dev)](https://travis-ci.org/alex023/eventbus?branch=dev)
[![Coverage Status](https://coveralls.io/repos/github/alex023/eventbus/badge.svg?branch=dev)](https://coveralls.io/github/alex023/eventbus?branch=dev)

## Brief
Eventbus is  event framework for in-memory event management. 
- Fault isolation
- Asynchronous event dispatching
- Filter support:enable log 、intercept、monitor messages.
- [todo]Multicasting events
- [todo]State monitor:enable monitor the topic statistic.
- [todo]Topic matchers:enable a subscriber to subscribe many topics

## Just Three Steps
1. define events
```golang
type message struct { /* Additional fields if needed */ }
```
We can ignore this step if basic type ,such as string、int、uint etc.. 
2. Prepare subscribers: Declare  your subscribing method without Mutex for multithreading,
```golang
//we can use any method name to  handle event 
func  HandleMsg(msg interface{}) {/* Do something */};

//...
func main(){
    eventbus.Default.Subscribe("topic",HandleMsg)
    //...
}
```
3. Push events
```golang
eventbus.Default.Push("topic",this is test msg!")
```
## Can I use it?
 The implementation is still in beta, we are using it for our production already. But the API change will be happen until 1.0.

## Examples
###  base function
this code show eventbus basic function：
1. fault isolation
2. the subscription object itself is unsubscribed
3. eventbus stop gracefull
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

//this code show eventbus basic func此代码演示消费者订阅的基本功能：
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
this code show how to use filter。therefor,we can log 、intercept、monitor etc with filter. 
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

//this code show how can use filter to intercept message！
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