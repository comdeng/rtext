package text

import (
	"log"
	"os"
	"strconv"
)

var fileIndex uint8 = 0
var filePos uint32 = 0

const TEXT_FILE_PREFIX = "text_"

// 写入文本内容
// 返回文件索引和所在文件位置
func Write(text []byte) (fIndex uint8, fPos uint32) {
	// Math.pow(2,32)
	if filePos > 4294967295 {
		fileIndex++
		filePos = 0
	}

	path := TEXT_FILE_PREFIX + strconv.Itoa(int(fileIndex)) + ".text"
	if filePos == 0 {
		log.Print(path)
		fc, _ := os.Create(path)
		fc.Close()
	}

	fh, _ := os.OpenFile(path, os.O_APPEND, 0)
	defer fh.Close()
	fh.WriteAt(text, int64(filePos))
	filePos += uint32(len(text))
	return
}

// 获取文本内容
func Read(fIndex uint8, fPos uint32, length uint16) (text []byte, ok bool) {
	path := TEXT_FILE_PREFIX + strconv.Itoa(int(fIndex)) + ".text"
	fo, err1 := os.Open(path)
	if err1 != nil {
		ok = false
		return
	}
	defer fo.Close()
	text = make([]byte, length)
	_, err2 := fo.ReadAt(text, int64(fPos))

	if err2 != nil {
		ok = false
		return
	}
	ok = true
	return
}
