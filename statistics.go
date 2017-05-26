package eventbus

import "sync"

type Statistics interface {
	TopicSubscribe(topicName string)
	TopicUnscribe(topicName string)
	MessagePushed(topicName string, message interface{})
	Topics() map[string]TopicInfo
}
type TopicInfo struct {
	Name      string
	Messages  int
	Consumers int
}
type DefaultStatistics struct {
	mut       sync.Mutex
	consumers map[string]int
	messages  map[string]int
}

func NewDefaultStatistics() Statistics {
	return &DefaultStatistics{
		consumers: make(map[string]int),
		messages:  make(map[string]int),
	}
}
func (ds *DefaultStatistics) TopicSubscribe(topicName string) {
	ds.mut.Lock()
	ds.consumers[topicName] = ds.consumers[topicName] + 1
	ds.mut.Unlock()
}
func (ds *DefaultStatistics) TopicUnscribe(topicName string) {
	ds.mut.Lock()
	ds.consumers[topicName] = ds.consumers[topicName] - 1
	ds.mut.Unlock()
}
func (ds *DefaultStatistics) MessagePushed(topicName string, message interface{}) {
	ds.mut.Lock()
	ds.messages[topicName] = ds.messages[topicName] + 1
	ds.mut.Unlock()
}

//Topics get every topic and it's input messages
func (ds *DefaultStatistics) Topics() map[string]TopicInfo {
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
