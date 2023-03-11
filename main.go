package main

import (
	"net"

	log "github.com/sirupsen/logrus"
	"snorba.art/mail/inbound/smtp"
)

func main() {
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1025")
	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		log.Info("got connection")

		go smtp.HandleIncoming(conn, nil)
	}
}
