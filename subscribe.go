package eventbus

import "sync"

type Subscribe interface {
	//remove all message handle from subscribed topics
	UnscribeAll()
	//remove  message handle from subscribed topic
	// will get error if this topic is not subscribed
	Unscribe(topicName string) error
	//get subscribed topics
	Topics() []string
}

type defaultSubscribe struct {
	mut    sync.RWMutex
	bus    *Bus
	topics map[string]uint64 //topic-id
}

func (sub *defaultSubscribe) UnscribeAll() {
	sub.mut.Lock()
	defer sub.mut.Unlock()

	for topicname, id := range sub.topics {
		sub.bus.unsubscribe(topicname, id)
	}
	sub.topics = nil
}
func (sub *defaultSubscribe) Topics() []string {
	var result []string
	sub.mut.Lock()
	defer sub.mut.Unlock()
	for topic, _ := range sub.topics {
		result = append(result, topic)
	}
	return result
}
func (sub *defaultSubscribe) Unscribe(topicName string) error {
	sub.mut.Lock()
	defer sub.mut.Unlock()

	seq, founded := sub.topics[topicName]
	if !founded {
		return ErrUnvaliableTopic
	}
	sub.bus.unsubscribe(topicName, seq)
	delete(sub.topics, topicName)
	return nil
}
