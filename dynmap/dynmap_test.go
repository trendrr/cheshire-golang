package dynmap

import (
	"log"
	"testing"
)

//go test -v github.com/trendrr/goshire/cheshire
func TestJsonMarshal(t *testing.T) {
	mp := New()
	mp.PutWithDot("this.that.test", 80)
	mp.PutWithDot("this.eight", 8)
	bytes, _ := mp.MarshalJSON()
	log.Printf("Got JSON %s", bytes)

	un := New()
	un.UnmarshalJSON(bytes)

	unbytes, _ := mp.MarshalJSON()
	if string(unbytes) != string(bytes) {
		t.Errorf("JSon marshal failed (%s) != (%s)", unbytes, bytes)
	}
}

func TestURLEncode(t *testing.T) {
	mp := New()
	mp.PutWithDot("this.that.test", 80)
	mp.PutWithDot("this.eight", 8)
	url, err := mp.MarshalURL()
	if err != nil {
		t.Errorf("Error in url %s", err)
	}

	log.Printf("Got URL : %s", url)

	un := New()
	un.UnmarshalURL(url)

	if un.MustInt("this.that.test", 0) != 80 {
		t.Errorf("Unmarshal URL failure ")
	}
}

