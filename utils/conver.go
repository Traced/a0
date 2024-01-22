package utils

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"log"
	"regexp"
	"strconv"
	"unsafe"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func BytesToString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func StringToBytes(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&str))
}

func StringMapToBufferReader(m map[string]string) io.Reader {
	var buf bytes.Buffer
	for k, v := range m {
		buf.WriteString(k + "=" + v + "&")
	}
	return &buf
}

// SVGToPNG svg必须包含 width 和 height属性
func SVGToPNG(svg []byte) (pngContent []byte) {
	wh := regexp.MustCompile(`(width|height)="(\d+)"`).
		FindAllStringSubmatch(BytesToString(svg), 2)
	if len(wh) != 2 {
		return
	}
	w, h := ForceInt64(wh[0][2]), ForceInt64(wh[1][2])
	icon, _ := oksvg.ReadIconStream(bytes.NewReader(svg))
	icon.SetTarget(0, 0, float64(w), float64(h))
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)
	var buf bytes.Buffer
	if err := png.Encode(io.Writer(&buf), rgba); err != nil {
		log.Println("编码成png失败：", err)
	}
	return buf.Bytes()
}

func ForceInt64(i string) int {
	r, _ := strconv.Atoi(i)
	return r
}
