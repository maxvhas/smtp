package smtp

import (
	"net"

	log "github.com/sirupsen/logrus"
)

func Listen(host string) (net.Listener, error) {
	laddr, err := net.ResolveTCPAddr("tcp", host)
	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Error(err)
				continue
			}

			go HandleIncoming(conn)
		}
	}()

	return l, nil
}
