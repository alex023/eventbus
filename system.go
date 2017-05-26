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
	consumerSeq uint64
	stopFlag    int32
	state       *Statistics
}

// New create a eventbus
func New() *Bus {
	bus := &Bus{
		topics: make(map[string]*Topic),
		state:  boundStatistics(),
	}
	return bus
}

//Subscribe 订阅主题，要确保输入的clientId唯一，避免不同客户端注册的时候采用同样的ClientId，否则会被替换。
func (bus *Bus) Subscribe(topicName string, handle CallFunc) (*Subscribe, error) {
	if atomic.LoadInt32(&bus.stopFlag) == _CLOSED {
		return nil, ErrClosed
	}

	bus.rwmut.RLock()
	ch, found := bus.topics[topicName]
	bus.rwmut.RUnlock()
	if !found {
		ch = NewTopic(topicName)
		bus.rwmut.Lock()
		bus.topics[topicName] = ch
		bus.rwmut.Unlock()
	}
	seq := atomic.AddUint64(&bus.consumerSeq, 1)
	msg := cmdAddConsumer{
		topic:       topicName,
		consumerSeq: seq,
		callFn:      handle,
	}
	ch.PostCmdMessage(msg)

	if bus.state != nil {
		bus.state.TopicSubscribe(topicName)
	}

	return &Subscribe{topic: topicName, id: seq, bus: bus}, nil
}

//unsubscribe 取消订阅。
//	因为采用异步消息，可以在订阅的消息handule中调用Unsubscribe来注销该主题。
func (bus *Bus) unsubscribe(topicName string, id uint64) error {
	if atomic.LoadInt32(&bus.stopFlag) == _CLOSED {
		return ErrClosed
	}

	ch, found := bus.topics[topicName]
	if !found {
		return ErrUnvaliableTopic
	}
	msg := cmdRmConsumer{
		consumerSeq: id,
	}
	ch.PostCmdMessage(msg)

	if bus.state != nil {
		bus.state.TopicUnscribe(topicName)
	}
	return nil
}

// Push  push a message to topic mailbox
func (bus *Bus) Push(topicName string, message interface{}) error {
	if atomic.LoadInt32(&bus.stopFlag) == _CLOSED {
		return ErrClosed
	}

	bus.rwmut.RLock()
	ch, found := bus.topics[topicName]
	bus.rwmut.RUnlock()
	if !found {
		return ErrUnvaliableTopic
	}

	ch.PostUserMessage(message)

	if bus.state != nil {
		bus.state.MessagePushed(topicName, message)
	}
	return nil
}

func (bus *Bus) LoadFilter(topic string, filter Filter) error {
	if atomic.LoadInt32(&bus.stopFlag) == _CLOSED {
		return ErrClosed
	}

	bus.rwmut.RLock()
	ch, found := bus.topics[topic]
	bus.rwmut.RUnlock()
	if !found {
		return ErrUnvaliableTopic
	}

	ch.PostCmdMessage(cmdLoadFilter{
		topic:  topic,
		filter: filter,
	})
	return nil
}

func (bus *Bus) UnloadFilter(topic string, filter Filter) error {
	if atomic.LoadInt32(&bus.stopFlag) == _CLOSED {
		return ErrClosed
	}

	bus.rwmut.RLock()
	ch, found := bus.topics[topic]
	bus.rwmut.RUnlock()
	if !found {
		return ErrUnvaliableTopic
	}

	ch.PostCmdMessage(cmdUnloadFilter{
		topic:  topic,
		filter: filter})

	return nil
}

// Stop immediately stop the execution of eventbus ,dropping all unreceived msg in topic mailbox
func (bus *Bus) Stop() {
	if atomic.CompareAndSwapInt32(&bus.stopFlag, _RUNNING, _CLOSED) {
		wg := &sync.WaitGroup{}
		wg.Add(len(bus.topics))
		cmdMsg := cmdStop{
			wg: wg,
		}
		for _, ch := range bus.topics {
			ch.PostCmdMessage(cmdMsg)
		}
	}
}

//StopGracefull block for stoping eventbus until all messages are received by the consumer
func (bus *Bus) StopGracefull() {
	if atomic.CompareAndSwapInt32(&bus.stopFlag, _RUNNING, _CLOSED) {
		wg := &sync.WaitGroup{}
		wg.Add(len(bus.topics))
		cmdMsg := cmdStopGracefull{
			wg: wg,
		}
		for _, ch := range bus.topics {
			ch.PostCmdMessage(cmdMsg)
		}
		wg.Wait()
	}
}

func (bus *Bus) Topics() map[string]TopicInfo {
	return bus.state.Topics()
}
