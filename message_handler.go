package smtp

type Envelope struct {
	Source      string
	Destination []string
	Body        []byte
}

type MessageHandler interface {
	Handle(Envelope)
}
