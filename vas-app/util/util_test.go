package util

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"qiniu.com/vas-app/biz/proto"
)

func TestJsonStr(t *testing.T) {
	str := JsonStr(map[string]int{"name": 1})
	assert.NotEmpty(t, str)
}

func TestFont(t *testing.T) {
	data, _ := ioutil.ReadFile("sample2.jpg")
	img, err := jpeg.Decode(bytes.NewReader(data))
	assert.Nil(t, err)
	c := image.NewRGBA(img.Bounds())
	draw.Draw(c, img.Bounds(), img, image.ZP, draw.Src)
	ret, err := DrawString(c, 1000, 1000, "饿了么", proto.MapTrafficDetectClassColor(3))
	assert.Nil(t, err)
	fout, err := os.Create("sample_out.jpg")
	assert.Nil(t, err)

	var b bytes.Buffer
	err = jpeg.Encode(&b, ret, &jpeg.Options{Quality: 100})
	assert.Nil(t, err)
	fout.Write(b.Bytes())
	fout.Close()
}
