package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"time"
)

type SlowActor struct {
}

func (c *SlowActor) Action(message interface{}) {
	time.Sleep(time.Second * 10)
}

// 此代码演示eventbus的两种关闭方式！
func main() {
	stop()
	stopGraceful()
}

func stop() {
	fmt.Println("we can use [Stop] to stop EventBus immediately.")
	eb := eventbus.New()
	consumer := &SlowActor{}

	eb.Subscribe("sleep", "consumer", consumer.Action)
	eb.Publish("sleep", struct{}{})
	eb.Stop()
}

func stopGraceful() {
	fmt.Println("we can use [StopGracefull] to stop EventBus until all consumer received message.")
	eb := eventbus.New()
	consumer := &SlowActor{}

	eb.Subscribe("sleep", "consumer", consumer.Action)
	eb.Publish("sleep", struct{}{})
	eb.StopGracefull()
}
