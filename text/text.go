package text

import (
	//"log"
	"os"
	"strconv"
)

var fileIndex uint8 = 0
var filePos uint32 = 0

const TEXT_FILE_PREFIX = "data/text_"

// 写入文本内容
// 返回文件索引和所在文件位置
func Write(text []byte) (fIndex uint8, fPos uint32) {
	// Math.pow(2,32)
	if filePos > 4294967295 {
		fileIndex++
		filePos = 0
	}

	path := getFilePath(fileIndex)
	if filePos == 0 {
		fc, _ := os.Create(path)
		fc.Close()
	}

	fh, _ := os.OpenFile(path, os.O_APPEND, 0)
	defer fh.Close()
	fh.WriteAt(text, int64(filePos))
	fPos = filePos

	//log.Printf("text.Write fileIndex=%d,filePos=%d,length=%d", fileIndex, filePos, len(text))

	filePos += uint32(len(text))
	return
}

// 获取文本内容
func Read(fIndex uint8, fPos uint32, length uint16) (text []byte, ok bool) {
	//log.Printf("text.Read fileIndex=%d,filePos=%d,length=%d", fIndex, fPos, length)

	path := getFilePath(fIndex)
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

func getFilePath(fileIndex uint8) string {
	return TEXT_FILE_PREFIX + strconv.Itoa(int(fileIndex)) + ".text"
}
