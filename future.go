package eventbus


type future struct{
	notify chan struct{}
}