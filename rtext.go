package main

import (
	_index "github.com/comdeng/rtext/index"
	_text "github.com/comdeng/rtext/text"
	"io"
	"log"
	"net/http"
	"strconv"
)

func main() {
	http.HandleFunc("/text/exists/", textExists)
	http.HandleFunc("/text/write/", textWrite)
	http.HandleFunc("/text/get/", textGet)

	// 构建索引的map
	_index.Build()

	err := http.ListenAndServe(":8889", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

// 文本是否存在
func textExists(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//log.Printf("/text/exists/?textId=%s", r.FormValue("textId"))

		textId, _ := strconv.ParseUint(r.FormValue("textId"), 10, 64)

		if textId < 1 {
			//log.Println("textExists.u_textId_illegal")
			io.WriteString(w, "-1")
		} else if _index.Exists(textId) {
			//log.Printf("exists")
			io.WriteString(w, "1")
		} else {
			//log.Printf("notexists")
			io.WriteString(w, "0")
		}
	}
}

func textGet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//log.Printf("/text/get/?textId=%s", r.FormValue("textId"))

		textId, _ := strconv.ParseUint(r.FormValue("textId"), 10, 64)

		if textId < 1 {
			//log.Println("textGet.u_textId_illegal")
			io.WriteString(w, "")
		} else {
			ii, ok := _index.Read(textId)
			if !ok {
				io.WriteString(w, "-1")
			}
			//log.Print(ii)
			if text, ok := _text.Read(ii.FileIndex, ii.FilePos, ii.Length); ok {
				io.WriteString(w, "1"+strconv.Itoa(int(ii.Flag)))
				io.WriteString(w, string(text))
			} else {
				io.WriteString(w, "0")
			}
		}
	}
}

// 文本写入
func textWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		flag, _ := strconv.Atoi(r.FormValue("flag"))
		if flag != 1 && flag != 0 {
			//log.Println("textWrite.u_flag_illegal")
			io.WriteString(w, "-1")
			return
		}
		text := r.FormValue("text")
		textId, _ := strconv.ParseUint(r.FormValue("textId"), 10, 64)

		//log.Printf("/text/write/?textId=%d&flag=%d", textId, flag)

		if textId > 1 && len(text) > 0 {
			if !_index.Exists(textId) {
				doTextWrite(textId, []byte(text), uint8(flag))
				io.WriteString(w, "1")
			}
		} else {
			io.WriteString(w, "0")
		}
	}
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
