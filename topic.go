package eventbus

import (
	"fmt"
	"github.com/alex023/eventbus/mailbox"
	"reflect"
	"sync/atomic"
)

//Topic 是一个订阅消息的运行对象定义，封装了邮箱和消费者
type Topic struct {
	Name      string
	mailbox   mailbox.Mailbox
	filters   map[Filter]bool
	consumers map[uint64]CallFunc
	msgCount  uint64
	closeFlag int32
}

// NewTopic topic constructor
func NewTopic(topicName string) *Topic {
	topic := &Topic{
		Name:      topicName,
		filters:   make(map[Filter]bool),
		consumers: make(map[uint64]CallFunc),
	}
	topic.mailbox = mailbox.New(1000, topic, mailbox.NewDispatcher(5))
	return topic
}

func (topic *Topic) addConsumer(consumerSeq uint64, handler CallFunc) {
	topic.consumers[consumerSeq] = handler
}

//rmConsumer remove callback function by assigned clientid
func (topic *Topic) rmConsumer(consumerSeq uint64) int {
	ret := len(topic.consumers)
	if _, founded := topic.consumers[consumerSeq]; founded {
		if ret > 0 {
			delete(topic.consumers, consumerSeq)
			ret--
		}
	}
	return ret
}
func (topic *Topic) PostCmdMessage(message interface{}) {
	if atomic.LoadInt32(&topic.closeFlag) == _CLOSED {
		return
	}

	topic.mailbox.PostCmdMessage(message)
}
func (topic *Topic) PostUserMessage(message interface{}) {
	if atomic.LoadInt32(&topic.closeFlag) == _CLOSED {
		return
	}

	topic.mailbox.PostUserMessage(message)
}

//notifyConsumer 向订阅了Topic的client发送消息
func (topic *Topic) notifyConsumer(message interface{}) {
	//i, t := 0, topic.dispatcher.Throughput()
	for _, handle := range topic.consumers {
		//if i > t {
		//	i = 0
		//	runtime.Gosched()
		//}
		//i++
		handle(message)
	}
}
func (topic *Topic) ReceiveUserMessage(message interface{}) {
	if msg, ok := message.(cmdStopGracefull); ok {
		topic.Close()
		msg.wg.Done()
		return
	}

	var proceed = true
	for filter, _ := range topic.filters {
		if proceed = filter.HandleMessage(message); !proceed {
			proceed = false
		}
	}
	if proceed {
		topic.notifyConsumer(message)
	}
}

func (topic *Topic) ReceiveCmdMessage(message interface{}) {
	switch msg := message.(type) {
	case cmdLoadFilter:
		topic.filters[msg.filter] = true
	case cmdUnloadFilter:
		delete(topic.filters, msg.filter)
	case cmdAddConsumer:
		topic.addConsumer(msg.consumerSeq, msg.callFn)
	case cmdRmConsumer:
		topic.rmConsumer(msg.consumerSeq)
	case cmdStop:
		topic.Close()
		msg.wg.Done()
	case cmdStopGracefull:
		topic.PostUserMessage(msg)
	default:
		fmt.Printf("[topic %v] receive undefined msg type:%v!\n", topic.Name, reflect.TypeOf(msg))
	}
}

// Close close topic until all messages have been sent to the registered client.
func (topic *Topic) Close() {
	if atomic.CompareAndSwapInt32(&topic.closeFlag, _RUNNING, _CLOSED) {
		topic.clean()
	}
}
func (topic *Topic) clean() {
	//release watcher & consumer
	//todo:should set nil after delete?
	for filter, _ := range topic.filters {
		delete(topic.filters, filter)
	}
	topic.filters = nil

	for id, _ := range topic.consumers {
		delete(topic.consumers, id)
	}
	topic.consumers = nil
}
