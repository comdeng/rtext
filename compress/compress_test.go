package compress

import "testing"

func TestCompress(t *testing.T) {
	orig := []byte("hello world")
	dist := Compress(orig)

	t.Log(orig, dist)
}

func TestUncompress(t *testing.T) {

}
