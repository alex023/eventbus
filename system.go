// Package eventbus is a simple subscription service module that provides asynchronous message distribution based on single computer memory
package eventbus

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrUnvaliableTopic = errors.New("unvaliable topic.")
	ErrClosed          = errors.New("closed eventbus.")
)

type CallFunc func(message interface{})

type cmdSubscribe struct {
	topic      string
	ConsumerId string
	callFunc   func(msg interface{})
}

type cmdUnsubscribe struct {
	topic      string
	ConsumerId string
}

type cmdLoadMiddleware struct {
	topic      string
	middleWare Filter
}
type cmdUnloadMiddleware struct {
	topic      string
	middleWare Filter
}
type cmdAddConsumer struct {
	topic      string
	ConsumerId string
	callFunc   func(message interface{})
}
type cmdRmConsumer struct {
	topic string
	id    string
}

type cmdStop struct{

}
// Bus is a  subscription service module
type Bus struct {
	rwmut sync.RWMutex
	dict  map[string]*Topic //map[topic.Name]*Channel
	//wg       basekit.WaitWraper
	msgCount uint64
	exitFlag int32
}

// New create a eventbus
func New() *Bus {
	bus := &Bus{
		dict: make(map[string]*Topic),
		//msgCache: make(chan *messageEnvelop, 1000),
	}
	//bus.p = NewProcess(bus)

	//bus.wg.Wrap(func() { bus.popMsg() })
	return bus
}

//Subscribe 订阅主题，要确保输入的clientId唯一，避免不同客户端注册的时候采用同样的ClientId，否则会被替换。

func (s *Bus) Subscribe(topic, clientID string, callFunc CallFunc) {
	s.rwmut.RLock()
	ch, found := s.dict[topic]
	s.rwmut.RUnlock()
	if !found {
		ch := NewTopic(topic)
		msg := cmdAddConsumer{
			topic:      topic,
			ConsumerId: clientID,
			callFunc:   callFunc,
		}
		ch.PostCmdMessage(msg)

		s.rwmut.Lock()
		s.dict[topic] = ch
		s.rwmut.Unlock()
	} else {
		msg := cmdAddConsumer{
			topic:      topic,
			ConsumerId: clientID,
			callFunc:   callFunc,
		}
		ch.PostUserMessage(msg)
	}
}

//Unsubscribe 取消订阅。
//	因为采用异步消息，可以在订阅的消息handule中调用Unsubscribe来注销该主题。
func (s *Bus) Unsubscribe(topicName string, clientID string) {
	ch, found := s.dict[topicName]

	if found {
		if ch.rmConsumer(clientID) == 0 {
			ch.Close()
			delete(s.dict, topicName)
		}
	}
}

// PostUserMessage asynchronous push a messageEnvelop
func (s *Bus) Publish(topic string, m interface{}) error {
	if atomic.LoadInt32(&s.exitFlag) == 1 {
		return ErrClosed
	}
	s.rwmut.RLock()
	ch, found := s.dict[topic]
	s.rwmut.RUnlock()
	if !found {
		return ErrUnvaliableTopic
	}
	ch.PostUserMessage(m)
	atomic.AddUint64(&s.msgCount, 1)
	return nil
}

func (s *Bus) LoadMiddle(topic string, ware Filter) {
	s.rwmut.RLock()
	ch, found := s.dict[topic]
	s.rwmut.RUnlock()
	if !found {
		return
	}
	ch.PostCmdMessage(cmdLoadMiddleware{
		topic:      topic,
		middleWare: ware,
	})
}

func (s *Bus) UnloadMiddle(topic string, ware Filter) {
	s.rwmut.RLock()
	ch, found := s.dict[topic]
	s.rwmut.RUnlock()
	if !found {
		return
	}
	ch.PostCmdMessage(cmdUnloadMiddleware{
		topic:      topic,
		middleWare: ware})
}

// Exiting returns a boolean indicating if topic is closed/exiting
func (s *Bus) Exiting() bool {
	return atomic.LoadInt32(&s.exitFlag) == 1
}

// Stop immediately stop the execution of eventbus ,dropping all unreceived msg in topic mailbox
func (s *Bus) Stop() {
	if atomic.CompareAndSwapInt32(&s.exitFlag, 0, 1) {
		//close(s.msgCache)
		//s.wg.Wait()
	}
}

//StopGracefull stop eventbus until all messages are received by the consumer
func (s Bus) StopGracefull() {
	//todo
}

//GetTopics performs a thread safe operation to get all topics in subscription service module
func (s *Bus) GetTopics() []string {
	result := make([]string, len(s.dict))
	i := 0
	for topic := range s.dict {
		result[i] = topic
		i++
	}
	return result
}
