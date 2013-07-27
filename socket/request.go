package socket

import (
	"bytes"
	//"log"
)

type Request struct {
	Url  string
	Data map[string]string
}

func (sr *Request) Decode(buf []byte) {
	b := bytes.NewBuffer(buf)

	url, _ := b.ReadString(0)
	url = url[0 : len(url)-1]
	sr.Url = url
	sr.Data = make(map[string]string)

	for {
		key, err := b.ReadString(0)
		if err != nil {
			panic(err)
		}
		if len(key) < 2 {
			break
		}
		key = key[0 : len(key)-1]

		lenBytes := make([]byte, 2)
		n, err := b.Read(lenBytes)
		if n < 2 || err != nil {
			break
		}
		length := uint16(lenBytes[0])<<8 | uint16(lenBytes[1])
		//log.Println(lenBytes)
		//log.Println(length)

		str := make([]byte, length)
		b.Read(str)

		sr.Data[key] = string(str)

		c, e := b.ReadByte()
		if e != nil || c == 0 {
			break
		} else {
			b.UnreadByte()
		}
	}

	//log.Println(sr.Url)
	//log.Println(sr.Data)
}
