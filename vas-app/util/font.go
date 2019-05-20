package util

import (
	"image"
	"image/color"
	"io/ioutil"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	log "qiniupkg.com/x/log.v7"
)

var ft *truetype.Font = nil

func GetFont(fontpath string) (*truetype.Font, error) {
	if ft != nil {
		return ft, nil
	}
	if fontpath == "" {
		fontpath = "./font.ttc"
	}
	fontBytes, err := ioutil.ReadFile(fontpath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	ft = f
	return ft, nil
}

func DrawString(img *image.RGBA, x, y int, str string, color color.Color) (*image.RGBA, error) {
	ft, err := GetFont("")
	if err != nil {
		return nil, err
	}
	c := freetype.NewContext()
	// c.SetDPI(72)
	c.SetFont(ft)
	c.SetFontSize(40)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color))

	pt := freetype.Pt(x, y)
	_, err = c.DrawString(str, pt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return img, nil
}
