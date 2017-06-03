package eventbus

type Subscribe interface {
	Unscribe()
	Topics() []string
}
type defaultSubscribe struct {
	bus    *Bus
	topics map[string]uint64 //topic-id
}

func (sub *defaultSubscribe) Unscribe() {
	for topicname, id := range sub.topics {
		sub.bus.unsubscribe(topicname, id)
	}
}
func (sub *defaultSubscribe) Topics() []string {
	var result []string
	for topic, _ := range sub.topics {
		result = append(result, topic)
	}
	return result
}
