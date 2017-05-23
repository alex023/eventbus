package mailbox

type Invoker interface {
	ReceiveCmdMessage(message interface{})
	ReceiveUserMessage(message interface{})
}
