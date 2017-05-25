// Package eventbus is a simple subscription service module that provides asynchronous message distribution based on single computer memory
package eventbus

import (
	"errors"
	"sync"
	"sync/atomic"
)

const (
	_CLOSED  = 1
	_RUNNING = 0
)

var (
	ErrUnvaliableTopic = errors.New("topic unvaliable.")
	ErrClosed          = errors.New("eventbus closed.")
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

type cmdLoadFilter struct {
	topic  string
	filter Filter
}
type cmdUnloadFilter struct {
	topic  string
	filter Filter
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

type cmdStop struct {
}
type cmdStopGracefull struct {
	wg *sync.WaitGroup
}

// Bus is a  subscription service module
type Bus struct {
	rwmut    sync.RWMutex
	topics   map[string]*Topic //map[topic.Name]*Channel
	msgCount uint64
	stopFlag int32
}

// New create a eventbus
func New() *Bus {
	bus := &Bus{
		topics: make(map[string]*Topic),
	}
	return bus
}

//Subscribe 订阅主题，要确保输入的clientId唯一，避免不同客户端注册的时候采用同样的ClientId，否则会被替换。
func (s *Bus) Subscribe(topicName, clientId string, callFn CallFunc) error {
	if atomic.LoadInt32(&s.stopFlag) == _CLOSED {
		return ErrClosed
	}

	s.rwmut.RLock()
	ch, found := s.topics[topicName]
	s.rwmut.RUnlock()
	if !found {
		ch = NewTopic(topicName)
		s.rwmut.Lock()
		s.topics[topicName] = ch
		s.rwmut.Unlock()
	}
	msg := cmdAddConsumer{
		topic:      topicName,
		ConsumerId: clientId,
		callFunc:   callFn,
	}
	ch.PostCmdMessage(msg)
	return nil
}

//Unsubscribe 取消订阅。
//	因为采用异步消息，可以在订阅的消息handule中调用Unsubscribe来注销该主题。
func (s *Bus) Unsubscribe(topicName string, clientId string) error {
	if atomic.LoadInt32(&s.stopFlag) == _CLOSED {
		return ErrClosed
	}

	ch, found := s.topics[topicName]

	if found {
		if ch.rmConsumer(clientId) == 0 {
			ch.Close()
			delete(s.topics, topicName)
		}
	}
	return nil
}

// PostUserMessage asynchronous push a messageEnvelop
func (s *Bus) Publish(topicName string, message interface{}) error {
	if atomic.LoadInt32(&s.stopFlag) == _CLOSED {
		return ErrClosed
	}

	s.rwmut.RLock()
	ch, found := s.topics[topicName]
	s.rwmut.RUnlock()
	if !found {
		return ErrUnvaliableTopic
	}

	ch.PostUserMessage(message)
	atomic.AddUint64(&s.msgCount, 1)
	return nil
}

func (s *Bus) LoadFilter(topic string, filter Filter) error {
	if atomic.LoadInt32(&s.stopFlag) == _CLOSED {
		return ErrClosed
	}

	s.rwmut.RLock()
	ch, found := s.topics[topic]
	s.rwmut.RUnlock()
	if !found {
		return ErrUnvaliableTopic
	}

	ch.PostCmdMessage(cmdLoadFilter{
		topic:  topic,
		filter: filter,
	})
	return nil
}

func (s *Bus) UnloadFilter(topic string, ware Filter) error {
	if atomic.LoadInt32(&s.stopFlag) == _CLOSED {
		return ErrClosed
	}

	s.rwmut.RLock()
	ch, found := s.topics[topic]
	s.rwmut.RUnlock()
	if !found {
		return ErrUnvaliableTopic
	}
	ch.PostCmdMessage(cmdUnloadFilter{
		topic:  topic,
		filter: ware})

	return nil
}

// Stop immediately stop the execution of eventbus ,dropping all unreceived msg in topic mailbox
func (s *Bus) Stop() {
	if atomic.CompareAndSwapInt32(&s.stopFlag, _RUNNING, _CLOSED) {
		//close(s.msgCache)
		//s.wg.Wait()
	}
}

//StopGracefull block for stoping eventbus until all messages are received by the consumer
func (s Bus) StopGracefull() {
	if atomic.CompareAndSwapInt32(&s.stopFlag, _RUNNING, _CLOSED) {
		wg := &sync.WaitGroup{}
		wg.Add(len(s.topics))

		cmdMsg := cmdStopGracefull{
			wg: wg,
		}
		for _, ch := range s.topics {
			ch.PostUserMessage(cmdMsg)
		}
		wg.Wait()
	}
}

//GetTopics performs a thread safe operation to get all topics in subscription service module
func (s *Bus) GetTopics() []string {
	result := make([]string, len(s.topics))
	i := 0
	for topic := range s.topics {
		result[i] = topic
		i++
	}
	return result
}
