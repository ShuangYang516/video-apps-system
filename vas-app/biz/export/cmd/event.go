package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/freetype"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	dashboardProto "qiniu.com/vas-app/biz/dashboard/dao/proto"
	baseProto "qiniu.com/vas-app/biz/proto"
	"qiniu.com/vas-app/util"
)

func ProcessEvents(list []baseProto.TrafficEventMsg, devices map[string]DeviceConfig, filePath string) (parentDir string, err error) {
	size := len(list)
	fmt.Println("events list size: ", size)
	if size == 0 {
		return
	}

	parentDir = path.Join(filePath, "export_"+time.Now().Format("20060102150405"))
	err = os.MkdirAll(parentDir, 0755)
	if err != nil {
		fmt.Printf("os.MkdirAll: %s, err: %s\n", parentDir, err)
		return
	}

	var (
		wg      sync.WaitGroup
		worker  = 100
		eventCh = make(chan baseProto.TrafficEventMsg, worker)
	)

	for i := 0; i < worker; i++ {
		wg.Add(1)
		go processEvent(&wg, eventCh, devices, parentDir)
	}

	for _, event := range list {
		if event.Status == baseProto.StatusFinished {
			eventCh <- event
		}
	}

	close(eventCh)
	wg.Wait()
	fmt.Println("all goroutine finish")

	return
}

func processEvent(wg *sync.WaitGroup, eventCh chan baseProto.TrafficEventMsg, devices map[string]DeviceConfig, parentDir string) {
	defer wg.Done()

	for event := range eventCh {
		doEvent(event, devices, parentDir)
	}
}

func doEvent(event baseProto.TrafficEventMsg, devices map[string]DeviceConfig, parentDir string) {
	fmt.Println("processEvent start")

	if len(event.Snapshot) == 0 {
		fmt.Println("event.Snapshot is null")
		return
	}

	device, ok := devices[event.CameraID]
	if !ok {
		device = DeviceConfig{}
	}

	eventNonMotor := event.EventType - baseProto.EventTypeNonMotorizedOffset
	if eventNonMotor > 0 && eventNonMotor < 100 {
		doNonMotorEvent(event, parentDir, device)
	} else if eventNonMotor > -100 && eventNonMotor < 0 {
		doVehicleEvent(event, parentDir, device)
	} else {
		fmt.Println("eventType is error")
	}

	fmt.Println("processEvent end")
}

func doNonMotorEvent(event baseProto.TrafficEventMsg, parentDir string, device DeviceConfig) {
	snapshort := event.Snapshot[0]
	faces := util.ProcessNonMotorDeepInfo(event.EventType, event.DeeperInfo)
	face := dashboardProto.People{}
	if len(faces) > 0 {
		face = faces[0]
	}

	huanhang := "\r\n"
	configs := "[TDRADAR]" + huanhang
	configs += "R_TYPE=" + "3" + huanhang

	now := time.Now()
	configs += "R_DATE=" + event.CreatedAt.Format("2006-01-02 15:04:05") + huanhang

	// todo addr
	configs += "R_ADDR=" + device.Addr + huanhang

	// todo code
	code := "9000154"
	configs += "R_CODE=" + code + huanhang

	configs += "R_NAME=" + face.Name + huanhang
	configs += "R_CARDID=" + face.IDCard + huanhang

	// todo
	phone := "17621555020"
	configs += "R_PHONE=" + phone + huanhang
	// todo
	configs += "R_SBBH=" + device.SBBH + huanhang

	// todo
	homeAddress := "伊犁路110弄4号103"
	configs += "R_HOMEADDRESS=" + homeAddress + huanhang
	// todo
	configs += "R_CARTYPE=" + "电动自行车" + huanhang
	class := baseProto.MapTrafficDetectClass(snapshort.Class)
	configs += "R_COMPANY=" + class + huanhang
	// todo
	configs += "R_INDUSTRY=" + "外卖" + huanhang
	eventType := baseProto.MapEventType(event.EventType)
	configs += "R_DESCRIPTION=" + eventType + huanhang

	dir := path.Join(parentDir, fmt.Sprintf("%s_%d", code, now.UnixNano()))
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("os.MkdirAll: %s, err: %v\n", dir, err)
		return
	}

	images := processImage(dir, &event, device, now.Format("20060102150405"))
	if len(images) != 2 {
		fmt.Println("len(images) != 2")
		os.RemoveAll(dir)
		return
	}

	// 下载人脸
	if face.FaceImageUrl != "" {
		yituFace, err := base64.StdEncoding.DecodeString(
			strings.TrimPrefix(face.FaceImageUrl, "data:image/jpg;base64,"))
		if err != nil {
			fmt.Printf("base64.StdEncoding.DecodeString(): %v\n", err)
		}
		out, err := os.Create(path.Join(dir, "face.jpg"))
		defer out.Close()
		_, err = io.Copy(out, bytes.NewReader(yituFace))
		if err != nil {
			fmt.Printf("io.Copy(): %v\n", err)
		}
	}

	configs += "R_TOTALFILE=" + strconv.Itoa(len(images)) + huanhang

	for index, image := range images {
		key := "R_FILE" + strconv.Itoa(index)
		configs += key + "=" + image

		if index < 1 {
			configs += huanhang
		}
	}

	bytes := []byte(configs)

	output, err := util.Utf8ToGbk(bytes)
	if err != nil {
		fmt.Println("convert to gbk err:", err)
		output = bytes
	}

	fileName := path.Join(dir, fmt.Sprintf("%s_%s.ini", code, now.Format("20060102150405")))
	err = ioutil.WriteFile(fileName, output, 0755)
	if err != nil {
		fmt.Printf("ioutil.WriteFile(%s): %v\n", fileName, err)
		return
	}
	// err = cfg.SaveToIndent(fileName, "\t")
	// if err != nil {
	// 	fmt.Printf("cfg.SaveToIndent(%s): %v\n", "", err)
	// 	return
	// }
}

func doVehicleEvent(event baseProto.TrafficEventMsg, parentDir string, device DeviceConfig) {
	snapshort := event.Snapshot[0]

	huanhang := "\r\n"
	configs := "[TDRADAR]" + huanhang
	configs += "R_TYPE=" + "1" + huanhang

	now := time.Now()
	configs += "R_DATE=" + event.CreatedAt.Format("2006-01-02 15:04:05") + huanhang

	// todo addr
	configs += "R_ADDR=" + device.Addr + huanhang

	// todo
	var codeH, codeT string
	labelRune := []rune(snapshort.Label)
	if len(labelRune) > 1 {
		codeH = string(labelRune[:1])
		codeT = string(labelRune[1:])
	}
	configs += "CODE_H=" + codeH + huanhang
	configs += "CODE_T=" + codeT + huanhang

	// todo  小型  中型  新能源
	configs += "R_CARTYPE=" + "小型车辆" + huanhang

	dir := path.Join(parentDir, fmt.Sprintf("%s_%d", snapshort.Label, now.UnixNano()))
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("os.MkdirAll: %s, err: %v\n", dir, err)
		return
	}

	images := processImage(dir, &event, device, now.Format("20060102150405"))
	if len(images) != 2 {
		fmt.Println("len(images) != 2")
		os.RemoveAll(dir)
		return
	}

	configs += "R_TOTALFILE=" + strconv.Itoa(len(images)) + huanhang

	for index, image := range images {
		key := "R_FILE" + strconv.Itoa(index)
		configs += key + "=" + image

		if index < 1 {
			configs += huanhang
		}
	}

	bytes := []byte(configs)

	output, err := util.Utf8ToGbk(bytes)
	if err != nil {
		fmt.Println("convert to gbk err:", err)
		output = bytes
	}

	fileName := path.Join(dir, fmt.Sprintf("%s_%s.ini", snapshort.Label, now.Format("20060102150405")))
	err = ioutil.WriteFile(fileName, output, 0755)
	if err != nil {
		fmt.Printf("ioutil.WriteFile(%s): %v\n", fileName, err)
		return
	}
}

func processImage(dir string, event *baseProto.TrafficEventMsg, device DeviceConfig, newNamePrefix string) (images []string) {
	images = make([]string, 0)
	snapshots := event.Snapshot

	if len(snapshots) < 3 {
		fmt.Println("len(snapshots) < 3")
		return
	}

	// todo
	text := []string{
		fmt.Sprintf("设备编号: %s 违法地点: %s", device.SBBH, device.Addr),
		// fmt.Sprintf("设备编号: %s 违法地点: %s", device.SBBH, device.Addr),
	}

	originImages := []string{
		snapshots[0].FeatureURI,
	}
	for _, snapshot := range snapshots {
		url := strings.Split(snapshot.SnapshotURI, ".jpg")[0] + "_raw.jpg"
		if snapshot.SnapshotURIRaw != "" {
			url = snapshot.SnapshotURIRaw
		}

		originImages = append(originImages, url)
	}

	count := 0
	for index, originImage := range originImages {
		if index >= 4 {
			break
		}
		if index%2 == 0 {
			continue
		}

		count++
		newName := fmt.Sprintf("%s-%d.jpg", newNamePrefix, count)
		newName, err := mergeImage(dir, originImages[index-1], originImage, newName, text)
		if err != nil {
			fmt.Printf("mergeImage(%s, %s, %s, %s): %v\n", dir, originImages[index-1], originImage, newName, err)
			continue
		}

		images = append(images, newName)
	}

	return
}

func mergeImage(dir string, url1 string, url2 string, newName string, text []string) (string, error) {
	src1, err := getImageObj(url1)
	if err != nil {
		return "", err
	}
	src1B := src1.Bounds().Max

	src2, err := getImageObj(url2)
	if err != nil {
		return "", err
	}
	src2B := src2.Bounds().Max

	newHeight := src1B.Y
	// resize 高度一致
	if src2B.Y > newHeight {
		newHeight = src2B.Y
		src1 = resize.Resize(0, uint(newHeight), src1, 100)
		src1B = src1.Bounds().Max
	} else if src2B.Y < newHeight {
		src2 = resize.Resize(0, uint(newHeight), src2, 100)
		src2B = src2.Bounds().Max
	}
	newWidth := src1B.X + src2B.X

	textHeight := 60
	des := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight+textHeight)) // 底板

	draw.Draw(des, des.Bounds(), src1, src1.Bounds().Min, draw.Over)                     //首先将一个图片信息存入jpg
	draw.Draw(des, image.Rect(src1B.X, 0, newWidth, src2B.Y), src2, image.ZP, draw.Over) //将另外一张图片信息存入jpg

	// 添加文字
	// Read the font data.
	var dpi, size float64 = 72, 36
	// var spacing = 1.5
	fontBytes, err := ioutil.ReadFile(fontfile)
	if err != nil {
		return "", err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return "", err
	}

	// Initialize the context.
	fg, bg := image.White, image.Black
	draw.Draw(des, image.Rect(0, newHeight, newWidth, newHeight+textHeight), bg, image.ZP, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(f)
	c.SetFontSize(size)
	c.SetClip(image.Rect(0, newHeight, newWidth, newHeight+textHeight))
	c.SetDst(des)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	pt := freetype.Pt(20, newHeight+10+int(c.PointToFixed(size)>>6))
	for index, s := range text {
		if index > 1 {
			break
		}

		_, err = c.DrawString(s, pt)
		if err != nil {
			fmt.Printf("c.DrawString(%s, %#v): %v\n", s, pt, err)
			continue
		}
		pt = freetype.Pt(20+src1B.X, newHeight+10+int(c.PointToFixed(size)>>6))
	}

	fSave, err := os.Create(dir + "/" + newName)
	if err != nil {
		return "", err
	}

	defer fSave.Close()

	var opt jpeg.Options
	opt.Quality = 80

	newImage := resize.Resize(3840, 0, des, resize.Lanczos3)

	err = jpeg.Encode(fSave, newImage, &opt) // put quality to 80%
	if err != nil {
		return "", err
	}

	return newName, nil
}

func getImageObj(url string) (img image.Image, err error) {
	data, err := util.Get(url)
	if err != nil {
		return
	}

	reader := bytes.NewReader(data)
	filetype := http.DetectContentType(data)

	switch filetype {
	case "image/jpeg", "image/jpg":
		img, err = jpeg.Decode(reader)
		if err != nil {
			fmt.Println("jpeg error")
			return nil, err
		}

	case "image/gif":
		img, err = gif.Decode(reader)
		if err != nil {
			return nil, err
		}

	case "image/png":
		img, err = png.Decode(reader)
		if err != nil {
			return nil, err
		}

	default:
		return nil, err
	}

	return img, nil
}
