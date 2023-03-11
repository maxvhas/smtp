package smtp

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"

	gosmtp "net/smtp"
)

type ServerTest struct {
	suite.Suite
	listener   net.Listener
	listenPort int
}

func (suite *ServerTest) SetupTest() {
	l, err := Listen("127.0.0.1:0")
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

	//fmt.Println("Done")
}

func (suite *ServerTest) TestMaxBodySize() {
	data := make([]byte, maxBodySize+1024*1024)

	_, err := rand.Read(data)
	if err != nil {
		suite.T().Error(err)
	}

	mailData := base64.RawStdEncoding.EncodeToString(data)

	err = gosmtp.SendMail(
		fmt.Sprintf("127.0.0.1:%d", suite.listenPort),
		nil,
		"sender@example.org",
		[]string{"receiver@example.org"},
		[]byte("To: receiver@example.com\r\nFrom: sender@example.com\r\n\r\n"+mailData+"\r\n.\r\n"),
	)

	if err == nil {
		suite.T().Fatal("Server did not close connection after data limit was exceeded.")
	}
}

func TestServerTest(t *testing.T) {
	suite.Run(t, new(ServerTest))
}
