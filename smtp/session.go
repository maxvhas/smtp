package smtp

import (
	"bufio"
	"bytes"
	"net"
	"regexp"

	log "github.com/sirupsen/logrus"
)

type session struct {
	conn    net.Conn
	scanner *bufio.Scanner

	respChan chan *Response
	quit     chan bool

	data []byte
	tls  bool
	ebm  bool

	client []byte
	src    []byte
	dst    []byte
	body   []byte
}

const maxBodyLines = 500

func newSession(conn net.Conn) *session {
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)
	return &session{
		conn:    conn,
		scanner: scanner,
		data:    make([]byte, 128),
	}
}

func HandleIncoming(conn net.Conn) {
	log.Info("incoming")
	s := newSession(conn)
	s.Greet()
	for {
		line, err := s.readLine()
		if err != nil {
			log.Error(err)
		}

		// EOF was reached, meaning the connection was closed
		if line == nil {
			break
		}

		log.Info("line:", string(line))

		cmd, err := ParseCommand(line)
		if err != nil {
			log.Error(err)
			// communicate parse error here
		}

		err = s.handleCommand(cmd)
		if err != nil {
			log.Error(err)
		}
	}
}

func (s *session) Greet() {
	_, err := s.conn.Write([]byte("220 localhost ESMTP Max\r\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func (s *session) readLine() (line []byte, err error) {
	if ok := s.scanner.Scan(); !ok {
		return nil, s.scanner.Err()
	}

	line = s.scanner.Bytes()
	s.data = append(s.data, line...)

	return line, nil
}

func (s *session) handleCommand(c *Command) error {
	switch c.code {
	case HELO:
		//do hello
		return s.HELO(c)
	case EHLO:
		return s.EHLO(c)
	case MAIL:
		return s.MAIL(c)
	case RCPT:
		return s.RCPT(c)
	case DATA:
		return s.DATA(c)
	case QUIT:
		break
	}

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

	// TODO: Check email validity
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
	s.dst = append(s.dst, matches[1]...)

	line := append([]byte("Recipient "), s.dst...)
	line = append(line, []byte(" ok")...)

	resp := Response{}
	resp.SetCode(RespOK)
	resp.AddLine(line)

	_, err := s.conn.Write(resp.Pack())
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
	for i := 0; i < maxBodyLines && s.scanner.Scan(); i++ {
		if err := s.scanner.Err(); err != nil {
			log.Error(err)
			resp.SetCode(RespFAILURE)
			resp.AddLine([]byte("Error while reading DATA segment"))
			s.conn.Write(resp.Pack())
			return err
		}

		line := s.scanner.Bytes()
		s.body = append(s.body, line...)
		if len(line) == 1 && bytes.Compare(line, []byte{'.'}) == 0 {
			resp.SetCode(RespOK)
			resp.AddLine([]byte("Ok: queued as a=@me"))
			s.conn.Write(resp.Pack())
			log.Info("Sent \"queued\" response")
			break
		} else {
			log.Info("DATA Line: " + string(line))
		}
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
