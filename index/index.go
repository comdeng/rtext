package index

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	TAG_Compressed      = 1
	TAG_NotCompressed   = 0
	INDEX_FILE_PREFIX   = "data/index_"
	INDEX_ROW_LENGTH    = 14
	INDEX_FILE_NUM      = 0x10
	TEXTID_BOUND_INDEX  = 0x8
	FILEPOS_BOUND_INDEX = 0xc
)

// 存10个索引
var indexPoses []uint64 = make([]uint64, INDEX_FILE_NUM)
var indexTree map[uint64]uint64 = make(map[uint64]uint64)

type IndexInfo struct {
	TextId    uint64
	Flag      uint8
	Length    uint16
	FileIndex uint8
	FilePos   uint32
}

// 将索引信息编码成14个字节的byte数组
func (ii *IndexInfo) Encode() (bytes []byte) {
	bytes = make([]byte, INDEX_ROW_LENGTH)

	bytes[0] = byte(ii.TextId >> 56)
	bytes[1] = byte(ii.TextId >> 48)
	bytes[2] = byte(ii.TextId >> 40)
	bytes[3] = byte(ii.TextId >> 32)
	bytes[4] = byte(ii.TextId >> 24)
	bytes[5] = byte(ii.TextId >> 16)
	bytes[6] = byte(ii.TextId >> 8)
	bytes[7] = byte(ii.TextId)

	bytes[8] = byte(ii.FilePos >> 24)
	bytes[9] = byte(ii.FilePos >> 16)
	bytes[10] = byte(ii.FilePos >> 8)
	bytes[11] = byte(ii.FilePos)

	bytes[12] = byte(ii.Length >> 1)
	bytes[13] = byte(byte(ii.Length&0x1)<<7 + ii.FileIndex<<1 + ii.Flag)

	return bytes
}

// 将索引信息解码
func (ii *IndexInfo) Decode(bytes []byte) {
	if ii == nil {
		ii = new(IndexInfo)
	}
	if len(bytes) != INDEX_ROW_LENGTH {
		panic("Length must be " + strconv.Itoa(INDEX_ROW_LENGTH))
	}

	ii.TextId = uint64(bytes[7]) | uint64(bytes[6])<<8 | uint64(bytes[5])<<16 | uint64(bytes[4])<<24 |
		uint64(bytes[3])<<32 | uint64(bytes[2])<<40 | uint64(bytes[1])<<48 | uint64(bytes[6])<<56

	ii.FilePos = uint32(bytes[11]) | uint32(bytes[10])<<8 | uint32(bytes[9])<<16 | uint32(bytes[8])<<24

	ii.Length = uint16(bytes[12]<<1 + bytes[13]>>7)
	ii.FileIndex = (bytes[13] & 0x7e) >> 1
	ii.Flag = bytes[13] & 0x0001
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

	bytes := ii.Encode()

	fh.WriteAt(bytes, int64(idxPos*INDEX_ROW_LENGTH))
	//log.Printf("index.Write textId=%d,idxIndex=%d,idxPos=%d,fileIndex=%d,filePos=%d,length=%d",
	//	ii.TextId,
	//	idxIndex, idxPos,
	//	ii.FileIndex, ii.FilePos, ii.Length)
	mapIndex(ii.TextId, idxPos)
	indexPoses[idxIndex]++
}

// 获取文本
func Read(textId uint64) (ii *IndexInfo, ok bool) {
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
	//log.Printf("index.Read textId=%d,idxIndex=%d,idexPos=%d,fileIndex=%d,filePos=%d,length=%d",
	//	textId, idxIndex, idxPos,
	//	ii.FileIndex, ii.FilePos, ii.Length,
	//)
	return
}

// 构建索引
func Build() {
	log.Println("index.Build start")
	start := time.Now()
	for i := uint8(0); i < INDEX_FILE_NUM; i++ {
		buildFileIndexes(i)
	}
	end := time.Now()
	log.Printf("index.Build end and cost %.2fms", float64(end.Sub(start)/1000000))
}

func buildFileIndexes(idxIndex uint8) {
	indexTree = make(map[uint64]uint64)
	path := getFilePath(idxIndex)
	fo, err := os.Open(path)
	if err != nil {
		return
	}

	idxPos := uint64(0)
	for {
		bytes := make([]byte, TEXTID_BOUND_INDEX)
		n, err := fo.Read(bytes)
		if err != nil || n == 0 {
			break
		}
		// 读出textId即可
		index := uint8(0)
		textId := uint64(0)
		for ; index < TEXTID_BOUND_INDEX; index++ {
			textId += (uint64(bytes[index]) << ((TEXTID_BOUND_INDEX - index - 1) << 3))
		}
		mapIndex(textId, idxPos)

		// 从当前位置前移6位
		fo.Seek(6, 1)

		idxPos++
	}
}

// 检查索引是否存在
func Exists(textId uint64) bool {
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
	return uint8(textId % INDEX_FILE_NUM)
}
