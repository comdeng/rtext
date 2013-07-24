package socket

import (
	"bytes"
	"log"
)

type Response struct {
	Status uint8
	Data   map[string]string
}

func (rs *Response) Encode() []byte {
	var b bytes.Buffer
	b.WriteByte(rs.Status)
	log.Print(rs.Data)
	for k, v := range rs.Data {
		b.WriteString(k)
		b.WriteByte(0)
		length := len(v)
		b.Write([]byte{uint8(length >> 8), uint8(length)})
		b.WriteString(v)
	}
	b.WriteByte(0)
	return b.Bytes()
}
