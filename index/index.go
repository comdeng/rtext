package index

import (
	"fmt"
	"os"
	"strconv"
)

const (
	TAG_Compressed    = 1
	TAG_NotCompressed = 0
	INDEX_FILE_PREFIX = "index_"
	INDEX_ROW_LENGTH  = 14
)

// 存10个索引
var indexPoses []uint64 = make([]uint64, 10)
var indexTree map[uint64]uint64 = make(map[uint64]uint64)

type IndexInfo struct {
	TextId    uint64
	Ctag      uint8
	Length    uint16
	FileIndex uint8
	FilePos   uint32
}

// 将索引信息编码成14个字节的byte数组
func (ii *IndexInfo) Encode() (bytes []byte) {
	bytes = make([]byte, 14)

	TextId := ii.TextId

	index := uint8(0)
	fmt.Printf("%x \n", ii.TextId)
	for ; index < 8; index++ {
		bytes[index] = byte(TextId >> ((8 - index - 1) * 8))
		fmt.Printf("%x \n", bytes[index])
	}
	fmt.Println(bytes)

	FilePos := ii.FilePos
	for ; index < 12; index++ {
		bytes[index] = byte(FilePos >> ((8 - index - 1) * 8))
	}

	bytes[12] = byte(ii.Length >> 1)
	bytes[13] = byte(byte(ii.Length&0x1)<<7 + ii.FileIndex<<1 + ii.Ctag)

	return bytes
}

// 将索引信息解码
func (ii *IndexInfo) Decode(bytes []byte) {
	if ii == nil {
		ii = new(IndexInfo)
	}
	if len(bytes) != 14 {
		panic("Length must be 14")
	}
	index := uint8(0)
	for ; index < 8; index++ {
		ii.TextId += (uint64(bytes[index]) << ((8 - index - 1) * 8))
	}

	for ; index < 12; index++ {
		ii.FilePos += uint32(bytes[index]) << ((14 - index - 1) * 8)
	}

	ii.Length = uint16(bytes[12]<<1 + bytes[13]>>7)
	fmt.Printf("%x ", bytes[13]>>1)
	ii.FileIndex = (bytes[13] & 0x7e) >> 1
	ii.Ctag = bytes[13] & 0x0001

}

// 检查TextId是否存在
func (ii *IndexInfo) Write() {
	idxIndex := getFileIndex(ii.TextId)
	idxPos := indexPoses[idxIndex]
	path := getFilePath(idxIndex)

	if idxPos == 0 {
		fo, _ := os.Create(path)
		fo.Close()
	}

	fh, _ := os.OpenFile(path, os.O_APPEND, 0)
	defer fh.Close()

	fh.WriteAt(ii.Encode(), int64(idxPos*INDEX_ROW_LENGTH))
	fmt.Println(ii)
	mapIndex(ii.TextId, idxPos)
	indexPoses[idxIndex]++
}

// 获取文本
func GetIndexInfo(textId uint64) (ii *IndexInfo, ok bool) {
	var idxPos uint64
	if idxPos, ok = indexTree[textId]; !ok {
		return
	}

	idxIndex := getFileIndex(textId)
	path := getFilePath(idxIndex)
	fo, err1 := os.Open(path)
	if err1 != nil {
		ok = false
		delete(indexTree, textId)
		return
	}
	defer fo.Close()
	bytes := make([]byte, INDEX_ROW_LENGTH)
	_, err2 := fo.ReadAt(bytes, int64(idxPos*INDEX_ROW_LENGTH))

	if err2 != nil {
		ok = false
		delete(indexTree, textId)
		return
	}
	ok = true
	ii = new(IndexInfo)
	ii.Decode(bytes)
	return
}

// 检查索引是否存在
func IndexExists(textId uint64) bool {
	_, ok := indexTree[textId]
	return ok
}

func mapIndex(textId uint64, pos uint64) {
	indexTree[textId] = pos
}

// 获取TextId对应的文件
func getFilePath(idxIndex uint8) string {
	return INDEX_FILE_PREFIX + strconv.Itoa(int(idxIndex)) + ".index"
}

func getFileIndex(textId uint64) uint8 {
	return uint8(textId % 10)
}
