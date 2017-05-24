package mailbox

import (
	"sync"
	"testing"
	"time"
)

const (
	producer_num    = 1000
	consumer_num    = 800
	action_times    = 1000
	action_duration = time.Second * 3
)

type producer struct {
	queue *Queue
	count int
}

func (p *producer) Action(times int) {
	for i := 0; i < times; i++ {
		p.queue.Push(struct{}{})
		p.count++
	}
}
func producerForEach(producers []producer, queue *Queue, producerActionTimes int) {
	for i := 0; i < len(producers); i++ {
		producers[i].queue = queue
	}

	var wg sync.WaitGroup
	wg.Add(len(producers))
	for _, p := range producers {
		go func() {
			p.Action(producerActionTimes)
			wg.Done()
		}()
	}
	wg.Wait()
}

type consumer struct {
	queue *Queue
	count int
}

func (c *consumer) Action(timeout time.Duration, maxReceive int) {
	t := time.After(timeout)
	for {
		select {
		case <-t:
			return
		default:
			if msg := c.queue.Pop(); msg != nil {
				c.count++

				if c.count >= maxReceive {
					return
				}
			}
		}
	}
}
func (c *consumer) Count() int {
	return c.count
}

func consumerForEach(consumers []consumer, queue *Queue, actionDuration time.Duration) {
	for i := 0; i < len(consumers); i++ {
		consumers[i].queue = queue
	}

	var wg sync.WaitGroup
	wg.Add(len(consumers))
	for i := 0; i < len(consumers); i++ {
		go func(i int) {
			consumers[i].Action(actionDuration, action_times)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestQueue_Push(t *testing.T) {
	queue := boundQueue(1000000)

	producers := make([]producer, producer_num)
	producerForEach(producers, queue, action_times)
}

func TestQueue_MulPushAndOnePop(t *testing.T) {
	queue := boundQueue(1000)

	ps := make([]producer, producer_num)
	go producerForEach(ps, queue, action_times)
	c := &consumer{
		queue: queue,
	}
	c.Action(time.Second, producer_num*action_times)

	if producer_num*action_times != c.Count() {
		t.Errorf("should=%v,actual=%v\n", producer_num*action_times, c.Count())
	}
}

func TestQueue_MulPushAndMulPop(t *testing.T) {
	queue := boundQueue(1000)

	ps := make([]producer, producer_num)
	go producerForEach(ps, queue, action_times)

	cs := make([]consumer, consumer_num)
	consumerForEach(cs, queue, action_duration)

	count := 0
	for _, c := range cs {
		count += c.Count()
	}
	if action_times*consumer_num != count {
		t.Errorf("should=%v,actual=%v\n", action_times*consumer_num, count)
	}
}
