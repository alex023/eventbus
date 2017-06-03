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
	ErrDuplicateTopic  = errors.New("topic duplicated.")
	ErrClosed          = errors.New("eventbus closed.")
)

type CallFunc func(message interface{})

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

var (
	defaultBus *Bus
	oncedo     sync.Once
)

//Default get default init Bus
func Default() *Bus {
	oncedo.Do(initDefaultBus)
	return defaultBus
}
func initDefaultBus() {
	defaultBus = New()
}

// Bus is a  subscription service module
type Bus struct {
	rwmut       sync.RWMutex
	topics      map[string]*Topic //map[topic.Name]*Channel
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

//Subscribe 订阅主题。
func (bus *Bus) Subscribe(handle CallFunc, topicNames ...string) (Subscribe, error) {
	var subscribe = &defaultSubscribe{
		topics: make(map[string]uint64),
		bus:    bus,
	}
	if atomic.LoadInt32(&bus.stopFlag) == _CLOSED {
		return nil, ErrClosed
	}
	for _, topicname := range topicNames {
		bus.rwmut.RLock()
		ch, found := bus.topics[topicname]
		bus.rwmut.RUnlock()
		if !found {
			ch = NewTopic(topicname, nil)
			bus.rwmut.Lock()
			bus.topics[topicname] = ch
			bus.rwmut.Unlock()
		}
		seq := atomic.AddUint64(&bus.consumerSeq, 1)
		msg := cmdAddConsumer{
			topic:       topicname,
			consumerSeq: seq,
			callFn:      handle,
		}
		ch.PostCmdMessage(msg)
		subscribe.topics[topicname] = seq
	}
	return subscribe, nil
}

//InitTopic use this method to set statics before Subscribe if we need to set
func (bus *Bus) InitTopic(topicName string, statics Statistics) error {
	if atomic.LoadInt32(&bus.stopFlag) == _CLOSED {
		return ErrClosed
	}

	bus.rwmut.RLock()
	ch, found := bus.topics[topicName]
	bus.rwmut.RUnlock()
	if !found {
		ch = NewTopic(topicName, statics)
		bus.rwmut.Lock()
		bus.topics[topicName] = ch
		bus.rwmut.Unlock()
	} else {
		return ErrDuplicateTopic
	}
	return nil
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

func (bus *Bus) Statistic(names ...string) []Statistics {
	var result []Statistics
	bus.rwmut.RLock()
	defer bus.rwmut.RUnlock()
	for _, topicname := range names {
		if topic, founded := bus.topics[topicname]; founded {
			result = append(result, topic.state)
		}
	}
	return result
}
