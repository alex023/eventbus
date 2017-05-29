package eventbus

import (
	"sync/atomic"
)

type Statistics interface {
	SetTopicName(topciName string)
	TopicSubscribe()
	TopicUnscribe()
	FilterLoad()
	FilterUnload()
	MessagePushed(message interface{})
	MessageReceived(message interface{})
}
type defaultStatistics struct {
	name            string
	subscribeNum    int64
	messageReceived uint64
	messagePushed   uint64
	filterNum       int64
}

func boundStatistics() Statistics {
	return &defaultStatistics{}
}
func (ds *defaultStatistics) SetTopicName(topicName string) {
	ds.name = topicName
}
func (ds *defaultStatistics) TopicSubscribe() {
	atomic.AddInt64(&ds.subscribeNum, 1)
}
func (ds *defaultStatistics) TopicUnscribe() {
	atomic.AddInt64(&ds.subscribeNum, -1)
}
func (ds *defaultStatistics) FilterLoad() {
	atomic.AddInt64(&ds.filterNum, 1)
}
func (ds *defaultStatistics) FilterUnload() {
	atomic.AddInt64(&ds.filterNum, -1)
}
func (ds *defaultStatistics) MessagePushed(message interface{}) {
	atomic.AddUint64(&ds.messagePushed, 1)
}

func (ds *defaultStatistics) MessageReceived(message interface{}) {
	atomic.AddUint64(&ds.messageReceived, 1)
}
