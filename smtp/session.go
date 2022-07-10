package smtp

import "net"

type Session struct {
	conn  net.Conn
	state *protocol.State
}
