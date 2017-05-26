package eventbus

import "sync"

type TopicInfo struct {
	Name      string
	Messages  int
	Consumers int
}

type Statistics struct {
	mut       sync.Mutex
	consumers map[string]int
	messages  map[string]int
}

func boundStatistics() *Statistics {
	return &Statistics{
		consumers: make(map[string]int),
		messages:  make(map[string]int),
	}
}
func (ds *Statistics) TopicSubscribe(topicName string) {
	ds.mut.Lock()
	ds.consumers[topicName] = ds.consumers[topicName] + 1
	ds.mut.Unlock()
}
func (ds *Statistics) TopicUnscribe(topicName string) {
	ds.mut.Lock()
	ds.consumers[topicName] = ds.consumers[topicName] - 1
	ds.mut.Unlock()
}
func (ds *Statistics) MessagePushed(topicName string, message interface{}) {
	ds.mut.Lock()
	ds.messages[topicName] = ds.messages[topicName] + 1
	ds.mut.Unlock()
}

//Topics get every topic and it's input messages
func (ds *Statistics) Topics() map[string]TopicInfo {
	result := make(map[string]TopicInfo)
	ds.mut.Lock()
	for topic, num := range ds.messages {
		result[topic] = TopicInfo{
			Name:      topic,
			Messages:  num,
			Consumers: ds.consumers[topic],
		}
	}
	ds.mut.Unlock()

	return result
}
