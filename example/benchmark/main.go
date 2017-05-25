package main

import (
	"fmt"
	"github.com/alex023/eventbus"
	"math/rand"
	"strconv"
	"time"
)

type ResultChan struct {
	signal chan int
}

type Consumer struct {
	ID      string
	counter int
}

func (c *Consumer) Count(message interface{}) {
	switch msg := message.(type) {
	case ResultChan:
		msg.signal <- c.counter
	default:
		c.counter++
	}
}

//此代码演示多消费者处理数据时的系统吞吐量！
func main() {
	//init [topic_num]topic,and arrange [consumer_num] consumer per topic。
	//send to [message_num] messages to every topic,should get 10*[message_num] messages。
	//So，we can get similar QPS！
	var (
		message_num  = 50000000
		topic_num    = 1000
		consumer_num = 10
	)
	eb := eventbus.New()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	resultChan := ResultChan{
		signal: make(chan int, topic_num),
	}

	//init 【topic_num】个主题，且为每个主题分配[consumer_num]个消费者
	for i := 0; i < topic_num; i++ {
		for j := 0; j < consumer_num; j++ {
			consumer := Consumer{
				ID: strconv.Itoa(i) + strconv.Itoa(j),
			}
			eb.Subscribe("subj"+strconv.Itoa(i), consumer.ID, consumer.Count)
		}
	}
	fmt.Printf("sending  %vw messages begin.......\n", message_num/10000)

	start := time.Now()
	//transmit ADD (by struct{}{}) signal
	for i := 0; i < message_num; i++ {
		if err := eb.Publish("subj"+strconv.Itoa(r.Intn(topic_num)), struct{}{}); err != nil {
			fmt.Println("publish err:", err)
			break
		}
	}
	//transmit EXIT (by ResultChan) signal
	for i := 0; i < topic_num; i++ {
		//NOTE:if transmit EXIT signal by other topic,it can be received before ADD ,and get uncorrect result
		eb.Publish("subj"+strconv.Itoa(i), resultChan)
	}
	var count = 0
	t := time.After(time.Second * 1)
	for {
		select {
		case result := <-resultChan.signal:
			count += result

		case <-t:
			goto OutLoop
		}
	}
OutLoop:
	end := time.Now()
	fmt.Printf("receive %vw messages by all consumers. \n", count/10000)
	fmt.Printf("spent：%2.0f seconds \n", end.Sub(start).Seconds())
}
