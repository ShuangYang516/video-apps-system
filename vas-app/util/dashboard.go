package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/qiniu/uuid"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gopkg.in/mgo.v2/bson"
	log "qiniupkg.com/x/log.v7"

	"qiniu.com/vas-app/biz/dashboard/dao/proto"
	dashboardProto "qiniu.com/vas-app/biz/dashboard/dao/proto"
	deepProto "qiniu.com/vas-app/biz/deeper/proto"
	baseProto "qiniu.com/vas-app/biz/proto"
)

func GenUuid() string {
	id, _ := uuid.Gen(16)
	return id
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToHZGB2312(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.HZGB2312.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

const (
	VmsLiveProtocolHFLV  = "hflv"
	VmsLiveProtocolHLS   = "hls"
	VmsLiveProtocolRTMP  = "rtmp"
	VmsLiveProtocolClose = "close"
)

func ProcessNonMotorDeepInfo(eventType int, deepInfo baseProto.DeeperInfo) (faces []dashboardProto.People) {
	eventNonMotor := eventType - baseProto.EventTypeNonMotorizedOffset
	if eventNonMotor <= 0 || eventNonMotor >= 100 {
		return
	}

	if deepInfo == nil {
		return
	}

	data, err := bson.Marshal(deepInfo)
	if err != nil {
		log.Error(err)
		return
	}

	deep := deepProto.DeeperInfo{}
	err = bson.Unmarshal(data, &deep)
	if err != nil {
		log.Error(err)
		return
	}

	if isFaceZero(deep.Face) || deep.Face.ErrorInfo != "" {
		return
	}

	if deep.Face.People == nil || len(deep.Face.People) == 0 {
		return
	}

	faces = make([]proto.People, len(deep.Face.People))
	for index, snapshortPeople := range deep.Face.People {
		if isPeopleZero(snapshortPeople) || len(snapshortPeople.Infos) == 0 {
			continue
		}

		info := snapshortPeople.Infos[0]

		if info.YituPeople == nil || len(info.YituPeople) == 0 {
			continue
		}

		similarPeople := info.YituPeople[0]

		faces[index] = proto.People{
			Name:         similarPeople.Name,
			Nation:       similarPeople.Nation,
			Sex:          similarPeople.Sex,
			IDCard:       similarPeople.IDCard,
			Similarity:   similarPeople.Similarity,
			FaceImageUrl: similarPeople.FaceImageUrl,
		}
	}

	return
}

func isFaceZero(face deepProto.DeeperFaceInfo) bool {
	return face.People == nil && face.ErrorInfo == ""
}

func isPeopleZero(people deepProto.DeeperFaceInfoArray) bool {
	return people.Infos == nil && people.FeatureURI == "" && people.FaceInfo == ""
}

func Zip(fw io.Writer, src string) (err error) {
	// 创建准备写入的文件
	// fw, err := os.Create(dst)
	// defer fw.Close()
	// if err != nil {
	// 	return err
	// }

	// 通过 fw 来创建 zip.Write
	zw := zip.NewWriter(fw)
	defer func() {
		// 检测一下是否成功关闭
		if err := zw.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// 下面来将文件写入 zw ，因为有可能会有很多个目录及文件，所以递归处理
	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) (err error) {
		if errBack != nil {
			return errBack
		}

		// 通过文件信息，创建 zip 的文件信息
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return
		}

		// 替换文件信息中的文件名
		fh.Name = strings.TrimPrefix(path, string(filepath.Separator))

		// 这步开始没有加，会发现解压的时候说它不是个目录
		if fi.IsDir() {
			fh.Name += "/"
		}

		// 写入文件信息，并返回一个 Write 结构
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return
		}

		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
		// 如目录，也没有数据需要写
		if !fh.Mode().IsRegular() {
			return nil
		}

		// 打开要压缩的文件
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return
		}

		// 将打开的文件 Copy 到 w
		n, err := io.Copy(w, fr)
		if err != nil {
			return
		}
		// 输出压缩的内容
		fmt.Printf("成功压缩文件： %s, 共写入了 %d 个字符的数据\n", path, n)

		return nil
	})
}
