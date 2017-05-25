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

// The Mailbox interface is used to enqueue messages to the mailbox
type Mailbox interface {
	PostUserMessage(message interface{})
	PostCmdMessage(message interface{})
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
		mailbox.dispatcher.Schedule(mailbox.processMsg)
	}
}
func (mailbox *defaultMailbox) processMsg() {
process_start:
	mailbox.run()

	atomic.StoreInt32(&mailbox.scheduleStatus, _IDLE)

	// check if there are still messages to process (sent after the message loop ended)
	user := atomic.LoadInt32(&mailbox.userMessages)
	cmd := atomic.LoadInt32(&mailbox.cmdMessages)
	if cmd > 0 || user > 0 {
		if atomic.CompareAndSwapInt32(&mailbox.scheduleStatus, _IDLE, _RUNNING) {
			goto process_start
		}
	}
}
func (mailbox *defaultMailbox) run() {
	var msg interface{}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("[EventBus] Recovering msg is ", msg, "Reson:", err)
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
			//Prioritize the cmdmessage
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

func New(size uint64, invoker Invoker, dispatcher Dispatcher) Mailbox {
	return &defaultMailbox{
		userMessage: boundQueue(size),
		cmdMessage:  boundQueue(100),
		invoker:     invoker,
		dispatcher:  dispatcher,
	}
}
