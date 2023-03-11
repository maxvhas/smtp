package smtp

import (
	"io"
	"sync"
)

type ResponseCode int

const (
	RespReady       ResponseCode = 220
	RespQuit        ResponseCode = 221
	RespOK          ResponseCode = 250
	RespStartMail   ResponseCode = 354
	RespFAILURE     ResponseCode = 554
	RespTooMuchData ResponseCode = 552
)

var ResponseCodeMap = map[ResponseCode][]byte{
	RespReady:       []byte("220"),
	RespQuit:        []byte("221"),
	RespOK:          []byte("250"),
	RespStartMail:   []byte("354"),
	RespFAILURE:     []byte("554"),
	RespTooMuchData: []byte("552"),
}

var SupportedExtensions = [][]byte{
	[]byte("PIPELINING"),
	[]byte("8BITMIME"),
}

// ConcurrentResponder is intended to be used by a single session
type ConcurrentResponder struct {
	w       io.Writer
	cmdChan chan Command
	wg      sync.WaitGroup
}

// Response has an SMTP-compliant status code and zero or more message lines
type Response struct {
	code  ResponseCode
	lines [][]byte
}

func (r *Response) SetCode(code ResponseCode) {
	r.code = code
}

func (r *Response) AddLine(line []byte) {
	r.lines = append(r.lines, line)
}

func (r *Response) Pack() []byte {
	var buf []byte
	numLines := len(r.lines)

	for i, line := range r.lines {
		var tmp []byte
		tmp = append(tmp, ResponseCodeMap[r.code]...)
		if i < numLines-1 {
			tmp = append(tmp, '-')
		} else {
			tmp = append(tmp, ' ')
		}

		tmp = append(tmp, line...)
		tmp = append(tmp, []byte{'\r', '\n'}...)
		buf = append(buf, tmp...)
	}

	return buf
}

func NewResponder(w io.Writer, c chan Command, wg sync.WaitGroup) *ConcurrentResponder {
	r := &ConcurrentResponder{
		w,
		c,
		wg,
	}

	return r
}

func (r *ConcurrentResponder) Start() {
	go r.handle()
}

func (r *ConcurrentResponder) handle() {

}
