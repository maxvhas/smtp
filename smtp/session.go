package smtp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"regexp"

	log "github.com/sirupsen/logrus"
)

type session struct {
	conn   net.Conn
	reader *bufio.Reader

	respChan chan *Response
	quit     chan bool
	done     bool

	config Config

	ebm bool

	client []byte
	src    []byte
	dst    [][]byte
	body   []byte
}

func (s *session) handleDelivery() {
	if s.config.MessageHandler != nil {

		destination := make([]string, len(s.dst))
		for i, dest := range s.dst {
			destination[i] = string(dest)
		}

		log.Info("Destinations:", destination)

		log.Info("Delegating message to message handler")
		s.config.MessageHandler.Handle(Envelope{
			Source:      string(s.src),
			Destination: destination,
			Body:        s.body,
		})
		return
	}

	log.Warn("No message handler configured, message discarded")
}

const maxBodySize = 56 * 1024 * 1024 // 56MB

type sessionReader struct {
	maxBodySize int
	readSize    int
	conn        net.Conn
}

var ErrBodySize = errors.New("Max size exceeded")

func (r *sessionReader) Read(b []byte) (int, error) {
	read, err := r.conn.Read(b)

	r.readSize += read
	if r.readSize > r.maxBodySize {
		return 0, ErrBodySize
	}

	return read, err
}

func newSessionReader(maxBodySize int, conn net.Conn) *sessionReader {
	return &sessionReader{
		maxBodySize: maxBodySize,
		conn:        conn,
	}
}

func newSession(conn net.Conn, config Config) *session {
	if config.MaxBodySize == 0 {
		config.MaxBodySize = maxBodySize
	}

	reader := newSessionReader(config.MaxBodySize, conn)

	return &session{
		conn:   conn,
		reader: bufio.NewReader(reader),
		config: config,
	}
}

func HandleIncoming(conn net.Conn, config Config) {
	log.Info("incoming")
	defer conn.Close()
	s := newSession(conn, config)
	err := s.Greet()
	if err != nil {
		log.Error(err)
		return
	}

	for !s.done {
		line, err := s.readLine()
		if err != nil {
			if errors.Is(err, ErrBodySize) {
				log.Info(err, ", Terminating session.")
				err = RespondTooMuchData(s.conn)
			} else {
				log.Error(err)
			}

			return
		}

		// EOF was reached, meaning the connection was closed
		if line == nil {
			log.Error("Unexpected behavior: Line was nil, but error was as well?", line)
			break
		}

		if len(line) == 0 {
			log.Warn("Zero-length line detected")
			continue
		}

		cmd, err := ParseCommand(line)
		if err != nil {
			log.Error(err)
			// communicate parse error here
			return
		}

		err = s.handleCommand(cmd)
		if err != nil {
			if errors.Is(err, ErrBodySize) {
				log.Info(err, ", Terminating session.")
				err = RespondTooMuchData(s.conn)
			} else {
				log.Error(err)
			}

			return
		}

		if s.done {
			log.Info("Session is DONE, terminating connection")
			break
		}
	}
}

func (s *session) Greet() error {
	_, err := s.conn.Write([]byte("220 localhost ESMTP Max\r\n"))

	return err
}

func (s *session) readLine() (line []byte, err error) {
	var part []byte
	var isPrefix bool

	for part, isPrefix, err = s.reader.ReadLine(); err == nil; part, isPrefix, err = s.reader.ReadLine() {
		line = append(line, part...)

		if !isPrefix {
			break
		}
	}

	// Prevent line from being nil when no error occured
	if err == nil && line == nil {
		line = make([]byte, 0)
	}

	return line, err
}

func (s *session) handleCommand(c *Command) error {
	switch c.code {
	case HELO:
		return s.HELO(c)
	case EHLO:
		return s.EHLO(c)
	case MAIL:
		return s.MAIL(c)
	case RCPT:
		return s.RCPT(c)
	case DATA:
		return s.DATA(c)
	case STARTTLS:
		return s.STARTTLS(c)
	case QUIT:
		return s.QUIT(c)
	}

	return nil
}

func (s *session) STARTTLS(c *Command) error {
	if !s.config.StartTLS {
		resp := Response{}
		resp.SetCode(RespTLSNotAvailable)
		resp.AddLine([]byte("TLS not available."))

		_, err := s.conn.Write(resp.Pack())
		return err
	}

	log.Info("Upgrading to TLS connection")
	client := tls.Server(
		s.conn,
		&tls.Config{
			Certificates: []tls.Certificate{
				*s.config.StartTLSCert,
			},
		},
	)

	resp := Response{}
	resp.SetCode(RespReady)
	resp.AddLine([]byte("Ready to start TLS"))

	_, err := s.conn.Write(resp.Pack())
	if err != nil {
		return err
	}

	log.Info("Performing TLS handshake")
	err = client.Handshake()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("Finished TLS handshake")

	// Replace tcp connection with the upgraded TLS one
	s.setConnection(client)

	// Discard any knowledge obtained prior the TLS upgrade as per RFC3207
	s.src = make([]byte, 0)
	s.dst = make([][]byte, 0)
	s.ebm = false
	s.body = make([]byte, 0)

	return nil
}

func (s *session) setConnection(conn net.Conn) {
	s.conn = conn
	s.reader = bufio.NewReader(newSessionReader(s.config.MaxBodySize, conn))
}

func (s *session) QUIT(c *Command) error {
	resp := Response{}
	resp.SetCode(RespQuit)
	resp.AddLine([]byte("OK"))
	_, err := s.conn.Write(resp.Pack())
	if err != nil {
		return err
	}

	s.done = true
	return nil
}

func (s *session) HELO(c *Command) error {
	resp := Response{}
	resp.SetCode(RespOK)
	resp.AddLine(append([]byte("Hello "), c.args...))

	_, err := s.conn.Write(resp.Pack())

	return err
}

func (s *session) EHLO(c *Command) error {
	resp := Response{}
	resp.SetCode(RespOK)
	resp.AddLine(append([]byte("Hello "), c.args...))

	for _, e := range SupportedExtensions {
		resp.AddLine(e)
	}
	_, err := s.conn.Write(resp.Pack())
	return err
}

var parseMAILArgsRegx = regexp.MustCompile(`(?i:From):\s*<([^>]+)>\s*(?i:BODY=([^\s]+))?`)

func (s *session) MAIL(c *Command) error {
	matches := parseMAILArgsRegx.FindSubmatch(c.args)
	if matches == nil {
		// error invalid MAIL syntax
	}

	if len(matches) != 3 {
		// error invalid MAIL syntax
	}

	// TODO: Check email validity (SPF)
	s.src = append(s.src, matches[1]...)

	if bytes.Compare(matches[2], []byte("8BITMIME")) == 0 {
		s.ebm = true
	}

	line := append([]byte("Sender "), s.src...)
	line = append(line, []byte(" ok")...)

	if s.ebm {
		line = append(line, []byte(" and 8BITMIME ok")...)
	}

	resp := Response{}
	resp.SetCode(RespOK)
	resp.AddLine(line)

	_, err := s.conn.Write(resp.Pack())
	return err
}

var parseRCPTArgsRegx = regexp.MustCompile(`(?i:To):\s*<([^>]+)>`)

// can we do this more generically?
func (s *session) RCPT(c *Command) error {
	matches := parseRCPTArgsRegx.FindSubmatch(c.args)
	//fmt.Println(matches)
	if matches == nil {
		// error invalid MAIL syntax
	}

	if len(matches) != 2 {
		// error invalid MAIL syntax
	}

	// TODO: Check email validity
	s.dst = append(s.dst, matches[1])

	line := append([]byte("Recipient "), matches[1]...)
	line = append(line, []byte(" ok")...)

	resp := Response{}
	resp.SetCode(RespOK)
	resp.AddLine(line)

	_, err := s.conn.Write(resp.Pack())
	return err
}

func RespondTooMuchData(conn net.Conn) error {
	resp := &Response{}
	resp.SetCode(RespTooMuchData)
	resp.AddLine([]byte("Too much data"))
	_, err := conn.Write(resp.Pack())

	return err
}

func (s *session) DATA(c *Command) error {
	resp := Response{}
	resp.SetCode(RespStartMail)
	resp.AddLine([]byte("End data with <CR><LF>.<CR><LF>"))
	log.Info("Executing DATA")
	_, err := s.conn.Write(resp.Pack())
	if err != nil {
		return err
	}

	log.Info("Reading DATA lines")
	resp = Response{}
	for line, err := s.readLine(); err == nil; line, err = s.readLine() {
		s.body = append(s.body, line...)
		s.body = append(s.body, '\r', '\n')

		if len(line) == 1 && bytes.Compare(line, []byte{'.'}) == 0 {
			resp.SetCode(RespOK)
			resp.AddLine([]byte("Ok: queued as a=@me"))
			s.conn.Write(resp.Pack())
			log.Info("Sent \"queued\" response")

			s.handleDelivery()
			break
		}
	}

	if err != nil && !errors.Is(err, io.EOF) {
		log.Error("Error reading data segment", err)
		resp.SetCode(RespFAILURE)
		resp.AddLine([]byte("Error while reading DATA segment"))
		s.conn.Write(resp.Pack())
		return err
	}

	log.Info("Finished reading DATA lines")

	return nil
}

func (s *session) startResponseWorker() {
	for {
		select {
		case <-s.quit:
			break
		case r := <-s.respChan:
			_, err := s.conn.Write(r.Pack())
			if err != nil {
				log.Error(err)
				s.quit <- true
				break
			}
		}
	}
}

// func (s *session) ReceiveCommand() error {
// 	// buf := make([]byte, command.MaxLen)

// 	s.ReadLine()
// 	// n, err := s.conn.Read(buf)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// fmt.Println(string(buf[0:n]))
// 	return nil
// }

// func (s *session) ReadLine() {
// 	sc := bufio.NewScanner(s.conn)
// 	sc.Split(split)

// 	for {
// 		if ok := sc.Scan(); !ok {
// 			break
// 		}
// 		fmt.Println(sc.Bytes())
// 	}
// }
