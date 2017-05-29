package mailbox

type Invoker interface {
	//优先调用的命令消息
	ReceiveCmdMessage(message interface{})
	//普通传递的消息
	ReceiveUserMessage(message interface{})
}
