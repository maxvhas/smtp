package smtp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"

	gosmtp "net/smtp"
)

type ServerTest struct {
	suite.Suite
	listener   net.Listener
	listenPort int
	handler    *messageHandlerWrapper
}

type messageHandlerWrapper struct {
	handleFunc func(Envelope)
}

func (w *messageHandlerWrapper) Handle(e Envelope) {
	if w.handleFunc != nil {
		w.handleFunc(e)
	}
}

func (suite *ServerTest) SetupTest() {
	suite.handler = &messageHandlerWrapper{}
	l, err := Listen("127.0.0.1:0", suite.handler)
	if err != nil {
		panic(err)
	}

	suite.listener = l
	suite.listenPort = l.Addr().(*net.TCPAddr).Port
}

func (suite *ServerTest) TestServer() {
	err := gosmtp.SendMail(
		fmt.Sprintf("127.0.0.1:%d", suite.listenPort),
		nil,
		"sender@example.org",
		[]string{"receiver@example.org"},
		[]byte("To: receiver@example.com\r\nFrom: sender@example.com\r\n\r\nbunzing aaa bunzing aaaaaa\r\n.\r\n"),
	)

	//	fmt.Println("Message: ", string([]byte("bunzing aaa bunzing aaaaaa\r\n.\r\n")))
	if err != nil {
		suite.T().Error(err)
	}
}

func (suite *ServerTest) TestMaxBodySize() {
	data := make([]byte, maxBodySize+1024*1024)

	_, err := rand.Read(data)
	if err != nil {
		suite.T().Error(err)
	}

	mailData := make([]byte, base64.RawStdEncoding.EncodedLen(len(data)))
	base64.RawStdEncoding.Encode(mailData, data)

	mailData = mailData[:len(data)]

	err = gosmtp.SendMail(
		fmt.Sprintf("127.0.0.1:%d", suite.listenPort),
		nil,
		"sender@example.org",
		[]string{"receiver@example.org"},
		[]byte("To: receiver@example.com\r\nFrom: sender@example.com\r\n\r\n"+string(mailData)+"\r\n.\r\n"),
	)

	if err == nil {
		suite.T().Fatal("Server did not close connection after data limit was exceeded.")
	}
}

func (suite *ServerTest) TestHandler() {
	var envelope Envelope
	suite.handler.handleFunc = func(e Envelope) {
		envelope = e
	}
	defer func() { suite.handler.handleFunc = nil }()

	data := make([]byte, maxBodySize-(1*1024*1024))

	_, err := rand.Read(data)
	if err != nil {
		suite.T().Error(err)
	}

	mailData := make([]byte, base64.RawStdEncoding.EncodedLen(len(data)))
	base64.RawStdEncoding.Encode(mailData, data)

	mailData = mailData[:len(data)]

	err = gosmtp.SendMail(
		fmt.Sprintf("127.0.0.1:%d", suite.listenPort),
		nil,
		"sender@example.org",
		[]string{"receiver@example.org", "receiver1@example.org", "receiver2@example.org"},
		[]byte("To: receiver@example.com\r\nFrom: sender@example.com\r\n\r\n"+string(mailData)+"\r\n.\r\n"),
	)

	if err != nil {
		suite.T().Error(err)
	}

	suite.Assert().Equal(
		[]string{"receiver@example.org", "receiver1@example.org", "receiver2@example.org"},
		envelope.Destination,
	)

}

func TestServerTest(t *testing.T) {
	suite.Run(t, new(ServerTest))
}
