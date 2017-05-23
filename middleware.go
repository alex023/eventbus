package eventbus

//Filter 是每个主题下在消费者接受到消息之前
type Filter interface {
	Receive(msg interface{}) (newMsg interface{}, proceed bool)
}

//创建一个基于责任链的消息执行函数
//func makeMiddleWareChain(middleware []Filter, lastReceiver Action) Action {
//	if len(middleware) == 0 {
//		return nil
//	}
//
//	h := middleware[len(middleware)-1](lastReceiver)
//	for i := len(middleware) - 2; i >= 0; i-- {
//		h = middleware[i](h)
//	}
//	return h
//}
