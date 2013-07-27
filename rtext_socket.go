package main

import (
	"flag"
	"fmt"
	_index "github.com/comdeng/rtext/index"
	_socket "github.com/comdeng/rtext/socket"
	_text "github.com/comdeng/rtext/text"
	"log"
	"net"
	"strconv"
	"time"
	//"strings"
)

const (
	STATUS_NOT_FOUND = 4
	STATUS_NO        = 3
	STATUS_YES       = 2
	STATUS_ERROR     = 1
)

type handler func(req *_socket.Request) *_socket.Response

var port *int = flag.Int("port", 8890, "Port on which to listen")
var handlers map[string]handler = map[string]handler{
	"/text/get":    handleGet,
	"/text/write":  handleWrite,
	"/text/exists": handleExists,
}

var handleId uint64 = 0

func waitForConnections(ls net.Listener) {
	for {
		conn, err := ls.Accept()
		if err != nil {
			log.Fatal("Got an error: %s", err)
		}

		go handleConnection(conn)
	}
}

// 处理连接
func handleConnection(conn net.Conn) {
	for {
		bs := make([]byte, 1024)
		n, _ := conn.Read(bs)
		if n == 0 {
			continue
		}
		if n == 1 && bs[0] == 1 {
			log.Print("close")
			conn.Close()
			break
		}

		req := new(_socket.Request)
		req.Decode(bs)

		res := new(_socket.Response)
		if h, ok := handlers[req.Url]; !ok {
			res.Status = STATUS_NOT_FOUND
		} else {
			handleId++
			log.Printf("[%d]:handle start:%s", handleId, req.Url)
			start := time.Now()
			res := h(req)
			buf := res.Encode()
			length := len(buf)
			conn.Write([]byte{uint8(length >> 8), uint8(length)})
			conn.Write(buf)
			end := time.Now()
			log.Printf("[%d]:handle end %s, cost %.2fms", handleId, req.Url, float64(end.Sub(start)/1000000))
			log.Print(buf)
			//conn.Close()
		}
	}
}

func main() {
	flag.Parse()
	ls, e := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if e != nil {
		log.Fatalf("Got an error: %s", e)
	}

	// 构建索引的map
	_index.Build()

	waitForConnections(ls)
}

// 处理获取文本的请求
func handleGet(req *_socket.Request) (ret *_socket.Response) {
	ret = new(_socket.Response)
	textId, _ := strconv.ParseUint(req.Data["textId"], 10, 64)
	if textId < 1 {
		ret.Status = STATUS_ERROR
	} else if ii, ok := _index.Read(textId); ok {
		if text, ok := _text.Read(ii.FileIndex, ii.FilePos, ii.Length); ok {
			ret.Status = STATUS_YES

			ret.Data = make(map[string]string)
			ret.Data["flag"] = strconv.Itoa(int(ii.Flag))
			ret.Data["text"] = string(text)
		} else {
			ret.Status = STATUS_NO
		}

	} else {
		ret.Status = STATUS_NOT_FOUND
	}
	return
}

// 写文本操作
func handleWrite(req *_socket.Request) (ret *_socket.Response) {
	ret = new(_socket.Response)
	textId, _ := strconv.ParseUint(req.Data["textId"], 10, 64)

	if textId > 1 && len(req.Data["text"]) > 0 {
		if !_index.Exists(textId) {
			flag, _ := strconv.Atoi(req.Data["flag"])
			doTextWrite(textId, []byte(req.Data["text"]), uint8(flag))
		}
		ret.Status = STATUS_YES
	} else {
		ret.Status = STATUS_NO
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

func handleExists(req *_socket.Request) (ret *_socket.Response) {
	ret = new(_socket.Response)
	textId, _ := strconv.ParseUint(req.Data["textId"], 10, 64)
	if textId < 1 {
		ret.Status = STATUS_ERROR
	} else if _index.Exists(textId) {
		ret.Status = STATUS_YES
	} else {
		ret.Status = STATUS_NOT_FOUND
	}
	return
}
