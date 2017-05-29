package mailbox

import (
	"sync"
	"testing"
	"time"
)

const (
	producer_num    = 1000
	consumer_num    = 500
	action_times    = 1000
	action_duration = time.Second
)

type producer struct {
	queue *Queue
	count int
}

func (p *producer) Action(times int) {
	for i := 0; i < times; i++ {
		p.queue.Post(struct{}{})
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

func (c *consumer) Action(maxReceive int, duration time.Duration) {
	t := time.After(duration)
	for {
		if msg := c.queue.Pop(); msg != nil {
			c.count++

			if c.count >= maxReceive {
				return
			}
		} else {
			select {
			case <-t:
				return
			default:
				//omit this select
			}
		}
	}
}
func (c *consumer) Count() int {
	return c.count
}

func consumerForEach(consumers []consumer, queue *Queue, duration time.Duration) {
	for i := 0; i < len(consumers); i++ {
		consumers[i].queue = queue
	}

	var wg sync.WaitGroup
	wg.Add(len(consumers))
	for i := 0; i < len(consumers); i++ {
		go func(i int) {
			consumers[i].Action(action_times, duration)
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
	c.Action(producer_num*action_times, action_duration)

	if producer_num*action_times != c.Count() {
		t.Errorf("should=%v,actual=%v\n", producer_num*action_times, c.Count())
	}
}

//TestQueue_MulPushAndMulPop  just test multithread safty.
func TestQueue_MulPushAndMulPop(t *testing.T) {
	//Because [RingBuffer] feature:" A put on full or get on empty call will block until an item
	// is put or retrieved.". So,[consumer_num] must less than [produce_num].

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
