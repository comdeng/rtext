package index

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	TAG_Compressed    = 1
	TAG_NotCompressed = 0
)

type indexInfo struct {
	//textId    uint64
	ctag      uint8
	length    uint32
	fileIndex uint8
	filePos   uint32
}

func DoEncode() []byte {
	ii := indexInfo{
		//	0x08090a0b,
		0x01,
		0x2a44,
		0x02,
		0x0001,
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, ii)
	return buf.Bytes()
}

func DoDecode() {
	b := []byte{
		//	0, 0, 0, 0, 8, 9, 10, 11,
		1,
		0, 0, 42, 68,
		2,
		0, 0, 0, 1,
	}
	var ii indexInfo

	binary.Read(bytes.NewBuffer(b), binary.BigEndian, &ii)

	fmt.Println(ii.fileIndex)
}
