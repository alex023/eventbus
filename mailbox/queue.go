package mailbox

import "github.com/Workiva/go-datastructures/queue"

type Queue struct {
	queue *queue.RingBuffer
}

//Pop get interface{} from queue
// This call will easy  blocked, if the queue is empty ,when multi goroutines
func (q *Queue) Pop() interface{} {
	if q.queue.Len() > 0 {
		item, _ := q.queue.Get()
		return item
	}
	return nil
}

func (q *Queue) Push(item interface{}) {
	q.queue.Put(item)
}

func boundQueue(buffer uint64) *Queue {
	return &Queue{
		queue: queue.NewRingBuffer(buffer),
	}
}
