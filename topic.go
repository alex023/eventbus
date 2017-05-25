package eventbus

import (
	"fmt"
	"github.com/alex023/eventbus/mailbox"
	"reflect"
	"sync/atomic"
)

//Topic 是一个订阅消息的运行对象定义，封装了邮箱和消费者
type Topic struct {
	Name       string
	mailbox    mailbox.Mailbox
	filters    map[Filter]bool
	consumers  map[string]CallFunc
	msgCount   uint64
	closeFlag  int32
	dispatcher mailbox.Dispatcher
}

// NewTopic topic constructor
func NewTopic(topicName string) *Topic {
	topic := &Topic{
		Name:       topicName,
		filters:    make(map[Filter]bool),
		consumers:  make(map[string]CallFunc),
		dispatcher: mailbox.NewDispatcher(5),
	}
	topic.mailbox = mailbox.New(1000, topic, topic.dispatcher)
	return topic
}

func (topic *Topic) addConsumer(clientId string, callFunc CallFunc) {
	topic.consumers[clientId] = callFunc
}

//rmConsumer remove callback function by assigned clientid
func (topic *Topic) rmConsumer(clientId string) int {
	ret := len(topic.consumers)

	if ret > 0 {
		delete(topic.consumers, clientId)
		ret--
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
func (topic *Topic) notifyConsumer(message interface{}) bool {
	//i, t := 0, topic.dispatcher.Throughput()
	for _, client := range topic.consumers {
		//if i > t {
		//	i = 0
		//	runtime.Gosched()
		//}
		//i++
		client(message)
	}
	return true
}
func (topic *Topic) ReceiveUserMessage(message interface{}) {
	if msg, ok := message.(cmdStopGracefull); ok {
		msg.wg.Done()
		return
	}

	var proceed bool
	for filter, _ := range topic.filters {
		if message, proceed = filter.Receive(message); !proceed {
			return
		}
	}
	topic.notifyConsumer(message)
}
func (topic *Topic) ReceiveCmdMessage(message interface{}) {
	switch msg := message.(type) {
	case cmdSubscribe:
		topic.addConsumer(msg.ConsumerId, msg.callFunc)
	case cmdUnsubscribe:
		topic.rmConsumer(msg.ConsumerId)
	case cmdLoadFilter:
		topic.filters[msg.filter] = true
	case cmdUnloadFilter:
		delete(topic.filters, msg.filter)
	case cmdAddConsumer:
		topic.addConsumer(msg.ConsumerId, msg.callFunc)
	case cmdRmConsumer:
		topic.rmConsumer(msg.id)
	case cmdStop:
		topic.Close()
	case cmdStopGracefull:
		topic.Close()
	default:
		fmt.Printf("[topic %v] receive undefined msg type:%v!\n", topic.Name, reflect.TypeOf(msg))
	}
}

// Close close topic until all messages have been sent to the registered client.
func (topic *Topic) Close() {
	if !atomic.CompareAndSwapInt32(&topic.closeFlag, _RUNNING, _CLOSED) {
		return
	}
	//等待正在执行的广播消息完成，通过wait确保注册方法的执行
}
