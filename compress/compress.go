package compress

import (
	"bytes"
	"compress/zlib"
	"fmt"
)

func Compress(orig []byte) (dist []byte) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(orig)
	w.Close()
	fmt.Println(b.Len())

	return b.Bytes()
}

func Uncompress(orig []byte, len int) (dist []byte) {
	b := bytes.NewBuffer(orig)
	r, err := zlib.NewReader(b)
	if err != nil {
		panic(err)
	}
	dist = make([]byte, len)
	r.Read(dist)
	return dist
}
