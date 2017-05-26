package eventbus

//Filter monitor the topic input message .one topic can add many Filter.
// the Filter  will be invoke before the topic send message to consumer.
type Filter interface {
	//HandleMessage handle message and return it's judge ,
	// if one or more Filter return FALSE ,the consumers can not receive it
	HandleMessage(message interface{}) (proceed bool)
}
