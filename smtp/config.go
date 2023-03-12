package smtp

import (
	"crypto/tls"
)

type Config struct {
	// Whether or not to support the STARTTLS command. Requires StartTLSCert to
	// be set.
	StartTLS     bool
	StartTLSCert *tls.Certificate

	// Max size of total connection body size, in bytes
	MaxBodySize int

	MessageHandler MessageHandler
}
