package mailbox

import "github.com/Workiva/go-datastructures/queue"

type Queue struct {
	queue *queue.RingBuffer
}

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

func BoundQueue(size uint64) *Queue {
	return &Queue{
		queue: queue.NewRingBuffer(size),
	}
}
