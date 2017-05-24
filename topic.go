package eventbus

import (
	"fmt"
	"github.com/alex023/eventbus/mailbox"
	"reflect"
	"runtime"
	"sync/atomic"
)

//Topic 是一个订阅消息的运行对象定义，封装了邮箱和消费者
type Topic struct {
	mailbox           mailbox.Mailbox
	Name              string
	consumers         map[string]CallFunc
	inBoundMidlewares map[Filter]bool
	msgCount          uint64
	closeFlag         int32
	dispatcher        mailbox.Dispatcher
}

// NewTopic topic constructor
func NewTopic(topicName string) *Topic {
	topic := &Topic{
		Name:              topicName,
		inBoundMidlewares: make(map[Filter]bool),
		consumers:         make(map[string]CallFunc),
		dispatcher:        mailbox.NewDispatcher(5),
	}
	topic.mailbox = mailbox.New(1000, topic, topic.dispatcher)
	return topic
}

func (topic *Topic) addConsumer(clientId string, callFunc CallFunc) {
	topic.consumers[clientId] = callFunc
}

//rmConsumer remove callback function by assigned clientid
func (topic *Topic) rmConsumer(clientID string) int {
	ret := len(topic.consumers)

	if ret > 0 {
		delete(topic.consumers, clientID)
		ret--
	}
	return ret
}
func (topic *Topic) PostCmdMessage(message interface{}) {
	topic.mailbox.PostCmdMessage(message)
}
func (topic *Topic) PostUserMessage(message interface{}) {
	topic.mailbox.PostUserMessage(message)
}

//notifyConsumer 向订阅了Topic的client发送消息
func (topic *Topic) notifyConsumer(message interface{}) bool {
	i, t := 0, topic.dispatcher.Throughput()
	for _, client := range topic.consumers {
		if i > t {
			i = 0
			runtime.Gosched()
		}
		i++
		client(message)
	}
	return true
}
func (topic *Topic) ReceiveUserMessage(message interface{}) {
	var proceed bool
	for middleware, _ := range topic.inBoundMidlewares {
		if message, proceed = middleware.Receive(message); !proceed {
			return
		}
	}
	topic.notifyConsumer(message)
}
func (topic *Topic) ReceiveCmdMessage(message interface{}) {
	switch msg := message.(type) {
	case cmdSubscribe:
		topic.addConsumer(msg.ConsumerId,msg.callFunc)
	case cmdUnsubscribe:
		topic.rmConsumer(msg.ConsumerId)
	case cmdLoadMiddleware:
		topic.inBoundMidlewares[msg.middleWare] = true
	case cmdUnloadMiddleware:
		delete(topic.inBoundMidlewares, msg.middleWare)
	case cmdAddConsumer:
		topic.addConsumer(msg.ConsumerId, msg.callFunc)
	case cmdRmConsumer:
		topic.rmConsumer(msg.id)
	case cmdStop:
		atomic.StoreInt32(&topic.closeFlag,1)
	default:
		fmt.Printf("[topic %v] receive undefined msg type:%v!\n", topic.Name, reflect.TypeOf(msg))

	}
}

// Close close mc topic until all messages have been sent to the registered client.
func (topic *Topic) Close() {
	if !atomic.CompareAndSwapInt32(&topic.closeFlag, 0, 1) {
		return
	}
	//等待正在执行的广播消息完成，通过wait确保注册方法的执行
}
