// Package eventbus is a simple subscription service module that provides asynchronous message distribution based on single computer memory
package eventbus

import (
	"sync"
	"sync/atomic"
)

type CallFunc func(message interface{})
type subscribe struct {
	topic    string
	clientId string
	callFunc func(msg interface{})
}

type unsubscribe struct {
	topic    string
	clientId string
}

type loadMiddleware struct {
	topic      string
	middleWare Filter
}
type unloadMiddleware struct {
	topic      string
	middleWare Filter
}
type addConsumer struct {
	topic    string
	id       string
	callFunc func(message interface{})
}
type removeConsumer struct {
	topic string
	id    string
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
		msg := addConsumer{
			topic:    topic,
			id:       clientID,
			callFunc: callFunc,
		}
		ch.PostCmdMessage(msg)

		s.rwmut.Lock()
		s.dict[topic] = ch
		s.rwmut.Unlock()
	} else {
		msg := addConsumer{
			topic:    topic,
			id:       clientID,
			callFunc: callFunc,
		}
		ch.PostUserMessage(msg)
	}
}

//Unsubscribe 取消订阅。由于内部使用了waitgroup，在使用时，要特别小心：
//	1.订阅某个主题的handle，其内部不得直接调用Unsubscribe来注销同一主题。否则，如果该主题正好只有最后一个client，就会被阻塞。
//	2.如果确实需要，请加入：关键字 go。
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
func (s *Bus) Publish(topic string, m interface{}) {
	if atomic.LoadInt32(&s.exitFlag) == 1 {
		return
	}
	s.rwmut.RLock()
	ch, found := s.dict[topic]
	s.rwmut.RUnlock()
	if !found {
		return
	}
	ch.PostUserMessage(m)
	atomic.AddUint64(&s.msgCount, 1)
}

func (s *Bus) LoadMiddle(topic string, ware Filter) {
	s.rwmut.RLock()
	ch, found := s.dict[topic]
	s.rwmut.RUnlock()
	if !found {
		return
	}
	ch.PostCmdMessage(loadMiddleware{
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
	ch.PostCmdMessage(unloadMiddleware{
		topic:      topic,
		middleWare: ware})
}

// Exiting returns a boolean indicating if topic is closed/exiting
func (s *Bus) Exiting() bool {
	return atomic.LoadInt32(&s.exitFlag) == 1
}

// Close safe exit eventbus
func (s *Bus) Close() {
	if atomic.CompareAndSwapInt32(&s.exitFlag, 0, 1) {
		//close(s.msgCache)
		//s.wg.Wait()
	}
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
