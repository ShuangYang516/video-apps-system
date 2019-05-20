package util

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"

	log "qiniupkg.com/x/log.v7"
)

type drawDot func(x, y int)

func DrawImageLines(imgData []byte, points []int) (ret []byte, err error) {
	if len(points)%2 != 0 {
		err = errors.New("invalid points intput")
		log.Error(err)
		return nil, err
	}
	img, err := jpeg.Decode(bytes.NewReader(imgData))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	lineImg := image.NewRGBA(img.Bounds())
	clr := color.RGBA{0, 255, 0, 255}
	for i, v := range points {
		if i >= 3 && i%2 == 1 {
			drawLine(points[i-3], points[i-2], points[i-1], v, func(x, y int) {
				lineImg.Set(x, y, clr)
			})
		}
	}

	c := image.NewRGBA(img.Bounds())
	draw.Draw(c, img.Bounds(), img, image.ZP, draw.Src)
	draw.Draw(c, img.Bounds(), lineImg, image.ZP, draw.Over)
	var b bytes.Buffer
	err = jpeg.Encode(&b, c, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return b.Bytes(), nil
}

func DrawImageLinesMulti(imgData []byte, pointsArr [][]int) (ret []byte, err error) {

	img, err := jpeg.Decode(bytes.NewReader(imgData))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	lineImg := image.NewRGBA(img.Bounds())
	clr := color.RGBA{0, 255, 0, 255}
	for _, points := range pointsArr {
		if len(points)%2 != 0 {
			err = errors.New("invalid points intput")
			log.Error(err)
			return nil, err
		}
		// points = append(points, points[0], points[1])
		for i, v := range points {
			if i >= 3 && i%2 == 1 {
				drawLine(points[i-3], points[i-2], points[i-1], v, func(x, y int) {
					lineImg.Set(x, y, clr)
				})
			}
		}
	}

	c := image.NewRGBA(img.Bounds())
	draw.Draw(c, img.Bounds(), img, image.ZP, draw.Src)
	draw.Draw(c, img.Bounds(), lineImg, image.ZP, draw.Over)

	var b bytes.Buffer
	err = jpeg.Encode(&b, c, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return b.Bytes(), nil
}

func drawLine(x0, y0, x1, y1 int, f drawDot) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	for {
		f(x0, y0)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}

	}
}

func abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}

func GetSnapshotFromStream(streamAddress string) (body []byte, err error) {
	imagename := ""
	sp := strings.Split(streamAddress, "/")
	if len(sp) <= 0 {
		err = errors.New("invalid streamAddress")
		return nil, err
	}
	imagename = fmt.Sprintf("%s-%d.jpg", sp[len(sp)-1], rand.Int())
	cmd := exec.Command("ffmpeg", "-i", streamAddress, "-r", "1", "-t", "1", "-f", "image2", imagename)

	output, err := cmd.CombinedOutput()
	log.Println(string(output))

	defer func() {
		if exist, _ := PathExists(imagename); exist {
			log.Println("remove temp file ", imagename)
			os.Remove(imagename)
		}
	}()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	body, err = ioutil.ReadFile(imagename)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return body, nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ExtractImage(imgData []byte, points []int) (ret []byte, err error) {
	if len(points) != 4 {
		err = errors.New("invalid points intput")
		log.Error(err)
		return nil, err
	}
	origin, fm, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	x0 := points[0]
	y0 := points[1]
	x1 := points[2]
	y1 := points[3]

	b := bytes.NewBuffer(make([]byte, 0))
	quality := 100
	switch fm {
	case "jpeg":
		img := origin.(*image.YCbCr)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.YCbCr)

		err = jpeg.Encode(b, subImg, &jpeg.Options{quality})
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return b.Bytes(), nil
	// case "png":
	// 	switch canvas.(type) {
	// 	case *image.NRGBA:
	// 		img := canvas.(*image.NRGBA)
	// 		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.NRGBA)
	// 		return png.Encode(out, subImg)
	// 	case *image.RGBA:
	// 		img := canvas.(*image.RGBA)
	// 		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
	// 		return png.Encode(out, subImg)
	// 	}
	// case "gif":
	// 	img := origin.(*image.Paletted)
	// 	subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.Paletted)
	// 	return gif.Encode(out, subImg, &gif.Options{})
	// case "bmp":
	// 	img := origin.(*image.RGBA)
	// 	subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
	// 	return bmp.Encode(out, subImg)
	default:
		return nil, errors.New("ERROR FORMAT")
	}
	return nil, nil
}

func LoadImage(uri string) ([]byte, error) {
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, errors.New("faceDetect连接出错:" + resp.Status)
		}
		defer resp.Body.Close()
		pix, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return pix, nil
	}
	return []byte(uri), nil
}
