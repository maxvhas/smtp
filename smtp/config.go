package smtp

import (
	"crypto/tls"
)

type Config struct {
	// Whether or not to support the STARTTLS command. Requires StartTLSCert to
	// be set.
	StartTLS     bool
	StartTLSCert *tls.Certificate

	MessageHandler MessageHandler
}
