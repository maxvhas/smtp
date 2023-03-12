package smtp

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"

	gosmtp "net/smtp"
)

var TestCertBlock = []byte(`-----BEGIN CERTIFICATE-----
MIIDkTCCAnkCFHzgm7BHX1qdPgp9fc9BqSvsNVS5MA0GCSqGSIb3DQEBCwUAMIGE
MQswCQYDVQQGEwJOTDEWMBQGA1UECAwNTm9vcmQtSG9sbGFuZDESMBAGA1UEBwwJ
QW1zdGVyZGFtMREwDwYDVQQKDAhMaWJyZSBJVDEUMBIGA1UEAwwLbGlicmUtaXQu
bmwxIDAeBgkqhkiG9w0BCQEWEWFkbWluQGxpYnJlLWl0Lm5sMB4XDTIzMDMxMjE1
MDExMVoXDTI1MTIwNzE1MDExMVowgYQxCzAJBgNVBAYTAk5MMRYwFAYDVQQIDA1O
b29yZC1Ib2xsYW5kMRIwEAYDVQQHDAlBbXN0ZXJkYW0xETAPBgNVBAoMCExpYnJl
IElUMRQwEgYDVQQDDAtsaWJyZS1pdC5ubDEgMB4GCSqGSIb3DQEJARYRYWRtaW5A
bGlicmUtaXQubmwwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDc78i3
DnxLtwDIC6peI9gzkvnjyHDXGNyN1CZ+RPu4cH90yTsh/sPzkPamd1UqNLY/hp/+
Pey0AMQm7oaRm+PhvsTU3Y4PMr1SMp9BOvVlFLvfHtiWTYuYuDxx0QZEGGt8kUyM
1ZCpRzS/lgkD55AeVpteEyPRilzG0AgXleN9ek4MaZbVMqqXhsIEOcGtsOs0VzND
QGSweps46fVdsbmr2ueJjNmXMSRxK/ihoHcUiaMuOkXPx7FlE+9BWQKkINGO3DUP
qsxv6WlXsyZI0q0jV/3+swo/CPvH/lo+iT/NYkodlQYkZmeZnK6E+L/FyqY8JFzd
xuRG5iW8uRfmrbf3AgMBAAEwDQYJKoZIhvcNAQELBQADggEBANLYdU7bhMRzRZNW
vRO+s5PdR6sR4E0qWkj3MA/LcQ6uL6LGdtdU0be/qG1iypRR3V3Jf+/sPCcCktKa
fDZwVthk21Pth4PoZk/KT6M8ztaO/kZG0tk29Sr4ynyxAhdHvGP4rZ8MdY9ZNDGZ
hyhoowvF2KXqn+ZNaUqO8uv/zzsxffm5WAff9piqoz4GPvwJwrXkX62b8Wg0Z8o2
b2rM6u9o/0Reqd51Rmw9mzW9TtGuJYaruWygsL8D/hlXzGYud35JSC+ioDg8ZZ8u
2hOwK0p/FE6fbLJM6mRdMbZLL2ULhE5womlQYsl3XCldg5l41kLfFeD2uJJWRDS1
t86cSAc=
-----END CERTIFICATE-----`)

var TestKeyBlock = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDc78i3DnxLtwDI
C6peI9gzkvnjyHDXGNyN1CZ+RPu4cH90yTsh/sPzkPamd1UqNLY/hp/+Pey0AMQm
7oaRm+PhvsTU3Y4PMr1SMp9BOvVlFLvfHtiWTYuYuDxx0QZEGGt8kUyM1ZCpRzS/
lgkD55AeVpteEyPRilzG0AgXleN9ek4MaZbVMqqXhsIEOcGtsOs0VzNDQGSweps4
6fVdsbmr2ueJjNmXMSRxK/ihoHcUiaMuOkXPx7FlE+9BWQKkINGO3DUPqsxv6WlX
syZI0q0jV/3+swo/CPvH/lo+iT/NYkodlQYkZmeZnK6E+L/FyqY8JFzdxuRG5iW8
uRfmrbf3AgMBAAECggEAKTUiAx6VCUwzPZyLZK6b0xa9PAJ1JXFSJbPlkBSOvJpi
82Xa/R627pVho6/LNymIunWCHtnu7a3c6AQCHmNsj/zUWn2OUwBcNloRwQldnsJM
vjNTI0mIWW43e+UIHahOV/gDxprItgH9cSRrPwqwIzB1HwlD23/KiRxg/gErYooV
4x3CIF9rQKmUSZM8fjrlGM4d/ue7oYHJjR8Tht2/6/7hOR1EiwtCuXBvnIsXhcZ/
rNmEBor7+PM/TsWAVSidyG4ChEnZLSVCvAEs45bm25nKf1za1JcZFz+klwThRRvm
XMcBmkg8jKBSzF3/9PJ7hOR8VdIuHjWfd7Z4djObBQKBgQD9UxSabuLy4p8GTF9i
bl6CP0HXghV2O8SrelvyDFKL70/hWLRtfMLvZdcrV4VmCc5KfLd1gvt3vj7YCysU
X2xmuTJZHrMHtlvUCH2nNyuly/mkgceU0RLcHH6G8oAepswZPaa5jcgFM5v6UlJH
h/IVy74iHK+oIGuHxzJWONkBAwKBgQDfRSK86vm+LSo6rEq11oYvdBQzQqJZZc0B
GQbnpH1kzTqA0v+ytPmW6G+5MeFt6DozauWTAh/UlLOOGy1U6HC1fxPyyL0YvSVp
+OSoV7ObMLGWJZxB+RwnI7AVAMwwYx6gXMwfTY+88xRJlQkVsbSS7GeAlUmgMCkC
BNuwNW/o/QKBgQCcO/gSAt9/UtsnBEUzrMQm6iKOalEYOWZjJ7S7RHRIj5Chd5bX
i8Gh6hpZRcIlG1kaQW7YT68Nu8yAa+rmxq9Rb1io9DEQSZy62X29el42A+X0WoIf
uw45qG00hy0TOmXYD1jbSaEZ7Cl/qfPK4AIjBSQ/X5fKRixrciQOX0MexwKBgQC2
vjadnFIHl54N4gFQbiLsaj0ya6LIOyudb2eYZ6j+vX/Z+1mwYrI7E0qGsU4LEF26
wg7f0YhODdwdPx9OdOXzl+yy9hzYR9B8uWwmYYovRp7D/0qzMPsbCfnQZxO5sxdZ
ODsWj/xLMkZzp5mE+SuMahSZSRe3FlQqQ+Gwizxq3QKBgQDZHQdJnJ86fKdY82Ol
ydzBxniars73nWqYZcjdxgzyof/PIjYPSX0ZPJ/ob3aqeSPd8hA1FwEO8ZIXtexK
sA8x7E74gonv1vWLV3fQc744wGiFFhFEhFdfqrbz474rq20AmrQFsD9HKE7s4Fat
IsbznQwGLLhIEfKOMG3RJE/T6Q==
-----END PRIVATE KEY-----`)

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

	certificate, err := tls.X509KeyPair(TestCertBlock, TestKeyBlock)
	if err != nil {
		suite.T().Error(err)
	}

	l, err := Listen(
		"127.0.0.1:0",
		Config{
			MessageHandler: suite.handler,
			StartTLS:       true,
			StartTLSCert:   &certificate,
		},
	)
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

func (suite *ServerTest) TestStartTLS() {
	client, err := gosmtp.Dial(fmt.Sprintf("127.0.0.1:%d", suite.listenPort))
	if err != nil {
		suite.T().Error(err)
	}

	// TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "localhost",
	}

	client.StartTLS(tlsConfig)

	err = client.Mail("sender@example.com")
	if err != nil {
		suite.T().Error(err)
	}

	err = client.Rcpt("receiver@example.com")
	if err != nil {
		suite.T().Error(err)
	}

	writer, err := client.Data()
	if err != nil {
		suite.T().Error(err)
	}

	_, err = writer.Write(
		[]byte(
			"To: receiver@example.com\r\n" +
				"From: sender@example.com\r\n" +
				"\r\nbunzing aaa bunzing aaaaaa\r\n.\r\n",
		),
	)
	if err != nil {
		suite.T().Error(err)
	}

	err = writer.Close()
	if err != nil {
		suite.T().Error(err)
	}

	err = client.Quit()
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
