package main

import "fmt"
import (
	//"bytes"
	//"compress/zlib"

	"github.com/comdeng/rtext/compress"
	"github.com/comdeng/rtext/index"
)

func main() {
	testBinary()
}

func testCompress() {
	orig := []byte(`17. js Format 提供JS格式化功能，快捷键 ctrl+alt+F，会根据
 
18. yui compressor 这个大家都知道yui的压缩工具，可以压缩CSS JS，直接CTRL+B，即可（需要安装配置了jdk之后才可用）
 
19. sublime v8 该插件提供jshint 及 v8引擎的js解析器console，jshint是JS语法校验器，较严格， v8则跟chrome里控制台一样。
 
20. ClipboardHistory： 该插件提供多剪贴板支持，你就可以同时保存多个剪贴板里的内容了，ctrl+alt+v快捷键调出`)
	b := compress.Compress(orig)
	fmt.Println(len(orig), len(b))
	fmt.Println(orig, b)
	fmt.Println(string(compress.Uncompress(b, len(orig))))
}

func testBinary() {
	//textId := 0x0000 ^ 10001
	//fmt.Printf("%x\n", textId)

	b := index.DoEncode()
	fmt.Println(b)
	index.DoDecode()
}
