package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"time"
)

type SlowActor struct {
	counter int
}

func (c *SlowActor) Action(message interface{}) {
	c.counter++
	fmt.Println("[SlowActor] count:", c.counter)
	time.Sleep(time.Second * 2)
}

// 此代码演示eventbus的两种关闭方式！
func main() {
	stop()
	time.Sleep(time.Second * 5)

	stopGraceful()
}

func stop() {
	fmt.Println("we can use [Stop] to stop EventBus immediately.")
	eb := eventbus.New()
	consumer := &SlowActor{}

	eb.Subscribe(consumer.Action, "sleep")
	eb.Push("sleep", struct{}{})
	eb.Push("sleep", struct{}{})
	eb.Stop()
}

func stopGraceful() {
	fmt.Println("we can use [StopGracefull] to stop EventBus until all consumer received message.")
	eb := eventbus.New()
	consumer := &SlowActor{}

	eb.Subscribe(consumer.Action, "sleep")
	eb.Push("sleep", struct{}{})
	eb.Push("sleep", struct{}{})
	eb.StopGracefull()
}
