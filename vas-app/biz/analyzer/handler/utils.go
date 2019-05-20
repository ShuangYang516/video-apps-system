package handler

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"time"

	log "qiniupkg.com/x/log.v7"
)

func waitForZeroSecond() {
	for {
		if time.Now().Second() == 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func abs(input int) int {
	if input < 0 {
		return -input
	}
	return input
}

//解析算法给出的原始gbr数据
func ConvertGBRData2Image(data []byte, width, height int) (img image.Image, err error) {
	if len(data) != width*height*3 {
		err = fmt.Errorf("len of data missmatch width and height")
		return nil, err
	}
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * 3
			r := data[offset+2]
			g := data[offset+1]
			b := data[offset]
			a := uint8(0)
			// c := color.RGBA{r, g, b, a}
			// rgba.Set(x, y, c)
			offset2 := rgba.PixOffset(x, y)
			rgba.Pix[offset2+0] = r
			rgba.Pix[offset2+1] = g
			rgba.Pix[offset2+2] = b
			rgba.Pix[offset2+3] = a
		}
	}
	return rgba, nil
}

func GetJpgData(img image.Image) (ret []byte, err error) {
	var buff bytes.Buffer
	err = jpeg.Encode(&buff, img, &jpeg.Options{Quality: 100})
	return buff.Bytes(), err
}

func ExtractJpgImage(jpgImage image.Image, points []int) (ret []byte, err error) {
	if len(points) != 4 {
		err = errors.New("invalid points intput")
		log.Error(err)
		return nil, err
	}

	x0 := points[0]
	y0 := points[1]
	x1 := points[2]
	y1 := points[3]

	b := bytes.NewBuffer(make([]byte, 0))
	quality := 100

	img := jpgImage.(*image.YCbCr)
	subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.YCbCr)

	err = jpeg.Encode(b, subImg, &jpeg.Options{quality})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return b.Bytes(), nil
}
