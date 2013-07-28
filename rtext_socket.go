package main

import (
	"flag"
	"fmt"
	_index "github.com/comdeng/rtext/index"
	_socket "github.com/comdeng/rtext/socket"
	_text "github.com/comdeng/rtext/text"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	//"time"
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
		s, err := ls.Accept()
		if err != nil {
			log.Fatal("Got an error: %s", err)
		}
		log.Printf("socket accept.%s", s.RemoteAddr().String())
		go handleConnection(s, s)
	}
}

// 处理连接
func handleConnection(conn net.Conn, r io.Reader) {
	defer func() {
		log.Print("socket closed")
		conn.Close()
	}()
	for {
		// time.Sleep(20 * time.Millisecond)
		//log.Print("for begin")
		bs := make([]byte, 4)

		n, err := r.Read(bs)
		//log.Printf("read n:%d", n)
		if err != nil {
			log.Printf("error:%s", err.Error())
			return
		}
		if n == 0 {
			//log.Print("receive 0 bit")
			continue
		}
		if n == 1 && bs[0] == 1 {
			return
		}
		//log.Print(bs)
		dataLen := uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3])
		//log.Printf("datLen is %d", dataLen)
		bs = make([]byte, int(dataLen))
		// 此处有时候数据很多，防止没有全部读取完毕
		n, _ = io.ReadFull(r, bs)
		//log.Printf("read data len:%d", n)
		if n == 0 {
			continue
		}

		req := new(_socket.Request)
		req.Decode(bs)
		//log.Printf("req.Url:%s", req.Url)

		res := new(_socket.Response)
		if h, ok := handlers[req.Url]; !ok {
			res.Status = STATUS_NOT_FOUND
		} else {
			handleId++
			//log.Printf("[%d]:handle start:%s", handleId, req.Url)
			//start := time.Now()
			res := h(req)
			buf := res.Encode()
			length := len(buf)
			conn.Write([]byte{uint8(length >> 8), uint8(length)})
			conn.Write(buf)
			//end := time.Now()
			//log.Printf("[%d]:handle end %s, cost %.2fms", handleId, req.Url, float64(end.Sub(start)/1000000))
			//log.Print(buf)
			//conn.Close()
		}
		//log.Print("for end")
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

var locker sync.RWMutex

// 写文本操作
func handleWrite(req *_socket.Request) (ret *_socket.Response) {
	ret = new(_socket.Response)
	textId, _ := strconv.ParseUint(req.Data["textId"], 10, 64)

	if textId > 1 && len(req.Data["text"]) > 0 {
		// 这一段内容需要锁定，保证唯一性
		locker.Lock()
		if !_index.Exists(textId) {
			flag, _ := strconv.Atoi(req.Data["flag"])
			doTextWrite(textId, []byte(req.Data["text"]), uint8(flag))
		}
		locker.Unlock()

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
	//log.Print(req.Data)
	if textId < 1 {
		ret.Status = STATUS_ERROR
	} else if _index.Exists(textId) {
		ret.Status = STATUS_YES
	} else {
		ret.Status = STATUS_NOT_FOUND
	}
	return
}
