package main

import (
	"log"

	"snorba.art/mail/inbound/smtp"
)

func main() {
	_, err := smtp.Listen("127.0.0.1:1025", smtp.Config{})
	log.Fatal(err)
}
