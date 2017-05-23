package mailbox

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

const (
	_IDLE int32 = iota
	_RUNNING
)

// The Mailbox interface is used to enqueue userMessages to the mailbox
type Mailbox interface {
	PostUserMessage(message interface{})
	PostCmdMessage(message interface{})
	Start()
}

type defaultMailbox struct {
	userMessage    *Queue //等待处理的消息队列
	userMessages   int32  //当前等待处理的消息数量
	cmdMessage     *Queue
	cmdMessages    int32
	invoker        Invoker    //消息处理器
	dispatcher     Dispatcher //并发器
	scheduleStatus int32      //调度标志，确定是否执行消息分配
}

func (mailbox *defaultMailbox) PostCmdMessage(message interface{}) {
	mailbox.cmdMessage.Push(message)
	atomic.AddInt32(&mailbox.cmdMessages, 1)
	mailbox.schedule()
}

func (mailbox *defaultMailbox) PostUserMessage(message interface{}) {
	mailbox.userMessage.Push(message)
	atomic.AddInt32(&mailbox.userMessages, 1)
	mailbox.schedule()
}

func (mailbox *defaultMailbox) schedule() {
	if atomic.CompareAndSwapInt32(&mailbox.scheduleStatus, _IDLE, _RUNNING) {
		mailbox.run()
		atomic.StoreInt32(&mailbox.scheduleStatus, _IDLE)
	}
}
func (mailbox *defaultMailbox) run() {
	var msg interface{}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("[EventBus] Recovering", "instance is ", mailbox.invoker, "Reson:", err)
		}
	}()

	i, t := 0, mailbox.dispatcher.Throughput()
	for {
		if i > t {
			i = 0
			runtime.Gosched()
		}
		i++
		if msg = mailbox.cmdMessage.Pop(); msg != nil {
			atomic.AddInt32(&mailbox.cmdMessages, -1)
			mailbox.invoker.ReceiveCmdMessage(msg)
			continue
		}

		if msg = mailbox.userMessage.Pop(); msg != nil {
			atomic.AddInt32(&mailbox.userMessages, -1)
			mailbox.invoker.ReceiveUserMessage(msg)
		} else {
			break
		}
	}

}
func (mailbox *defaultMailbox) Start() {

}
func New(size uint64, invoker Invoker, dispatcher Dispatcher) Mailbox {
	return &defaultMailbox{
		userMessage: BoundQueue(size),
		cmdMessage:  BoundQueue(5),
		invoker:     invoker,
		dispatcher:  dispatcher,
	}
}
