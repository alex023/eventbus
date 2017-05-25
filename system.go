// Package eventbus is a simple subscription service module that provides asynchronous message distribution based on single computer memory
package eventbus

import (
	"errors"
	"sync"
	"sync/atomic"
)

const (
	_RUNNING int32 = 0
	_CLOSED        = 1
)

var (
	ErrUnvaliableTopic = errors.New("topic unvaliable.")
	ErrClosed          = errors.New("eventbus closed.")
)

type CallFunc func(message interface{})

type Subscribe struct {
	bus   *Bus
	topic string
	id    uint64
}

func (sub *Subscribe) Unscribe() {
	sub.bus.unsubscribe(sub.topic, sub.id)
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
	topic       string
	consumerSeq uint64
	callFn      CallFunc
}
type cmdRmConsumer struct {
	consumerSeq uint64
}

type cmdStop struct {
	wg *sync.WaitGroup
}
type cmdStopGracefull struct {
	wg *sync.WaitGroup
}

// Bus is a  subscription service module
type Bus struct {
	rwmut       sync.RWMutex
	topics      map[string]*Topic //map[topic.Name]*Channel
	msgCount    uint64
	consumerSeq uint64
	stopFlag    int32
}

// New create a eventbus
func New() *Bus {
	bus := &Bus{
		topics: make(map[string]*Topic),
	}
	return bus
}

//Subscribe 订阅主题，要确保输入的clientId唯一，避免不同客户端注册的时候采用同样的ClientId，否则会被替换。
func (s *Bus) Subscribe(topicName string, handle CallFunc) (*Subscribe, error) {
	if atomic.LoadInt32(&s.stopFlag) == _CLOSED {
		return nil, ErrClosed
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
	seq := atomic.AddUint64(&s.consumerSeq, 1)
	msg := cmdAddConsumer{
		topic:       topicName,
		consumerSeq: seq,
		callFn:      handle,
	}
	ch.PostCmdMessage(msg)
	return &Subscribe{topic: topicName, id: seq, bus: s}, nil
}

//unsubscribe 取消订阅。
//	因为采用异步消息，可以在订阅的消息handule中调用Unsubscribe来注销该主题。
func (s *Bus) unsubscribe(topicName string, id uint64) error {
	if atomic.LoadInt32(&s.stopFlag) == _CLOSED {
		return ErrClosed
	}

	ch, found := s.topics[topicName]
	if !found {
		return ErrUnvaliableTopic
	}
	msg := cmdRmConsumer{
		consumerSeq: id,
	}
	ch.PostCmdMessage(msg)
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

func (s *Bus) UnloadFilter(topic string, filter Filter) error {
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
		filter: filter})

	return nil
}

// Stop immediately stop the execution of eventbus ,dropping all unreceived msg in topic mailbox
func (s *Bus) Stop() {
	if atomic.CompareAndSwapInt32(&s.stopFlag, _RUNNING, _CLOSED) {
		wg := &sync.WaitGroup{}
		wg.Add(len(s.topics))
		cmdMsg := cmdStop{
			wg: wg,
		}
		for _, ch := range s.topics {
			ch.PostCmdMessage(cmdMsg)
		}
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
			ch.PostCmdMessage(cmdMsg)
		}
		wg.Wait()
	}
}

//GetTopics performs a thread safe operation to get all topics in subscription service module
func (s *Bus) GetTopics() []string {
	s.rwmut.RLock()
	result := make([]string, len(s.topics))
	i := 0
	for topic := range s.topics {
		result[i] = topic
		i++
	}
	s.rwmut.RUnlock()
	return result
}
