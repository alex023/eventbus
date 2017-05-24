package mailbox

import "github.com/Workiva/go-datastructures/queue"

type Queue struct {
	queue *queue.RingBuffer
}
//Pop get interface{} from queue
// This call will easy  blocked, if the queue is empty ,when multi goroutines
func (mq *Queue) Pop() interface{} {
	if mq.queue.Len() > 0 {
		item, _ := mq.queue.Get()
		return item
	}
	return nil
}

func (mq *Queue) Push(item interface{}) {
	mq.queue.Put(item)
}

func boundQueue(buffer uint64) *Queue {
	return &Queue{
		queue: queue.NewRingBuffer(buffer),
	}
}
