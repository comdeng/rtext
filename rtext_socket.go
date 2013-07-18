package main

import (
	"bytes"
	"flag"
	"fmt"
	_index "github.com/comdeng/rtext/index"
	_text "github.com/comdeng/rtext/text"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	STATUS_NOT_FOUND = 4
	STATUS_NO        = 3
	STATUS_YES       = 2
	STATUS_ERROR     = 1
)

type handler func(req *rtextRequest) *rtextResponse

var port *int = flag.Int("port", 8890, "Port on which to listen")
var handlers map[string]handler = map[string]handler{
	"/text/get":    handleGet,
	"/text/write":  handleWrite,
	"/text/exists": handleExists,
}

type rtextRequest struct {
	url  string
	data map[string]string
}

func (rr *rtextRequest) Receive(r io.Reader, hdrBytes []byte) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)

	url, _ := buf.ReadString("\0")

	for {
		b := buf.ReadByte()
		if b != byte("\0") {
			key, _ := buf.ReadString(":")
			length, _ := strconv.buf.ReadString(":")
			data := make([]byte, length)
			buf.Read(data)
			rr.data[key] = string(data)
		} else {
			break
		}
	}
	rr.url = url
	rr.data = data
}

type rtextResponse struct {
	status int16
	data   []byte
}

type chanReq struct {
	req *rtextRequest
	res chan *rtextResponse
}

type reqHandler struct {
	ch chan chanReq
}

func (rh reqHandler) HandleMessage(w io.Writer, req *rtextRequest) *rtextResponse {
	cr := &chanReq{
		req,
		make(chan *rtextResponse),
	}
	rh.ch <- cr
	return <-cr.res
}

// Request handler for doing server stuff.
type RequestHandler interface {
	// Handle a message from the client.
	// If the message should cause the connection to terminate,
	// the Fatal flag should be set.  If the message requires no
	// response, return nil
	//
	// Most clients should ignore the io.Writer unless they want
	// complete control over the response.
	HandleMessage(io.Writer, *rtextRequest) *rtextResponse
}

func waitForConnections(ls net.Listener) {
	reqChannel := make(chan chanReq)

	go runServer(reqChannel)
	handler := &reqHandler{reqChannel}

	log.Printf("Listening on port %d", *port)
	for {
		s, e := ls.Accept()
		if e == nil {
			log.Printf("Got a connection from %v", s.RemoteAddr())
			go handleIO(s, handler)
		} else {
			log.Printf("Error accepting from %s", ls)
		}
	}
}

// Handle until the handler returns a fatal message or a read or write
// on the socket fails.
func handleIO(s io.ReadWriteCloser, handler reqHandler) error {
	defer s.Close()
	var err error
	for err == nil {
		err = handleMessage(s, s, handler)
	}
	return err
}

func handleMessage(r io.Reader, w io.Writer, handler RequestHandler) error {

}

func runServer(input chan chanReq) {
	for {
		req := <-input
		log.Printf("Got a request: %s", req.req)
		req.res <- dispatch(req.req)
	}
}

func dispatch(req *rtextRequest) (res *rtextResponse) {
	if h, ok := handlers[req.url]; !ok {
		res.status = STATUS_NOT_FOUND
	} else {
		h(req)
	}
	return
}

func main() {
	flag.Parse()
	ls, e := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if e != nil {
		log.Fatalf("Got an error: %s", e)
	}

	waitForConnections(ls)
}

// 处理获取文本的请求
func handleGet(req *rtextRequest) (ret *rtextResponse) {
	textId := uint64(strconv.Atoi(req.data["textId"]))
	if textId < 1 {
		ret.status = STATUS_ERROR
	} else if _index.Exists(textId) {
		ret.status = STATUS_YES
	} else {
		ret.status = STATUS_NOT_FOUND
	}
	return
}

// 写文本操作
func handleWrite(req *rtextRequest) (ret *rtextResponse) {
	textId := uint64(strconv.Atoi(req.data["textId"]))
	if textId > 1 && len(req.data["text"]) > 0 {
		if !_index.Exists(textId) {
			doTextWrite(textId, byte[](req.data["text"]), uint8(strconv.Atoi(req.data["flag"]))
			ret.status = STATUS_YES
		}
	} else {
		ret.status = STATUS_NO
	}
	return
}

// 写入文本内容
func doTextWrite(textId uint64, txt []byte, flag uint8) {
	if len(txt) > 65535 {
		panic("txt length is bigger than 65535")
	}
	length := uint16(len(txt))
	// 写文本
	fileIndex, filePos := _text.Write(txt)
	ii := &_index.IndexInfo{
		textId,
		flag,
		length,
		fileIndex,
		filePos,
	}

	// 写索引
	ii.Write()
}

func handleExists(req *rtextRequest) (ret *rtextResponse) {
	textId := uint64(strconv.Atoi(req.data["textId"]))
	if textId < 1 {
		ret.status = STATUS_ERROR
	} else if _index.Exists(textId) {
		ret.status = STATUS_YES
	} else {
		ret.status = STATUS_NO
	}
	return
}
