package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image"
	"image/jpeg"

	xlog "github.com/qiniu/xlog.v1"
	gproto "qiniu.com/vas-app/biz/proto"
	gutil "qiniu.com/vas-app/util"

	"qiniu.com/vas-app/biz/deeper/client"
	"qiniu.com/vas-app/biz/deeper/proto"
)

type FaceSearchHandlerConfig struct {
}

type FaceSearchHandler struct {
	FaceSearchHandlerConfig
	qiniuClient *client.QiniuClient
	yituClient  *client.YituClient
}

func NewFaceSearchHandler(config FaceSearchHandlerConfig, qiniuClient *client.QiniuClient, yituClient *client.YituClient) *FaceSearchHandler {
	return &FaceSearchHandler{
		FaceSearchHandlerConfig: config,
		qiniuClient:             qiniuClient,
		yituClient:              yituClient,
	}
}

func (h *FaceSearchHandler) Handle(ctx context.Context, input proto.TrafficEvent) (output proto.DeeperInfo) {
	xl := xlog.FromContextSafe(ctx)
	// 此处不确定判断标准, 暂定为小于30（30以上为违章）
	if input.EventType <= 30 {
		return
	}

	for _, snapshot := range input.Snapshot {
		if snapshot.FeatureURI == "" {
			// 如果没有featureURI, 则跳过
			continue
		}
		facesInSnapshot := proto.DeeperFaceInfoArray{}

		featureEncodedURI, facesDetected, err := h.yituFaceDetect(ctx, snapshot)
		if err != nil {
			xl.Errorf("faceDetect error: %+v", err)
			output.Face.ErrorInfo = "人脸检测失败"
			return output
		}
		facesInSnapshot.FeatureURI = snapshot.FeatureURI
		for _, face := range facesDetected {
			pts := face.PTS
			if len(facesDetected) == 1 {
				pts = [4][2]int{}
			}
			yituPeople, err := h.faceSearch(ctx, featureEncodedURI, pts)
			if err != nil {
				xl.Errorf("faceSearch error: %+v", err)
				output.Face.ErrorInfo = "人脸搜索失败"
				return output
			}
			face.YituPeople = yituPeople
			facesInSnapshot.Infos = append(facesInSnapshot.Infos, face)
		}

		// TODO 人脸照片其余情况待定，暂时只考虑未检索到人脸情况
		if facesInSnapshot.Infos == nil {
			facesInSnapshot.FaceInfo = "未检索到人脸"
		}
		output.Face.People = append(output.Face.People, facesInSnapshot)
	}
	return output
}

// 人脸检测 (qiniu版本)
func (h *FaceSearchHandler) qiniuFaceDetect(ctx context.Context, snapshot gproto.Snapshot, threshold float64) (encodedRawImage []byte, faces []proto.DeeperFacePeopleInfo, err error) {
	if h.qiniuClient == nil {
		err = errors.New("qiniuClient not initialized")
		return
	}
	xl := xlog.FromContextSafe(ctx)
	decodedURI, err := gutil.LoadImage(snapshot.FeatureURI)
	if err != nil {
		xl.Errorf("qiniuFaceDetect.LoadImage: %+v", err)
		return
	}
	encodedRawImage = []byte(base64.StdEncoding.EncodeToString(decodedURI))

	qiniuFaceDetectResult, err := h.qiniuClient.DetectFace(ctx, decodedURI)
	if err != nil {
		xl.Errorf("qiniuFaceDetect.DetectFace: %+v", err)
		return
	}
	for _, detection := range qiniuFaceDetectResult.Result.Detections {
		if detection.Score < threshold {
			continue
		}
		encodedFaceURI, err := clipImage(decodedURI, detection.Pts)
		if err != nil {
			xl.Errorf("qiniuFaceDetect.clipImage: %+v", err)
			return nil, nil, err
		}

		faces = append(faces, proto.DeeperFacePeopleInfo{
			PTS:         detection.Pts,
			Score:       detection.Score,
			Orientation: detection.Orientation,
			Quality:     detection.Quality,
			FaceUri:     encodedFaceURI,
		})
	}
	return
}

// 人脸检测 (yitu版本)
func (h *FaceSearchHandler) yituFaceDetect(ctx context.Context, snapshot gproto.Snapshot) (encodedRawImage []byte, faces []proto.DeeperFacePeopleInfo, err error) {
	if h.yituClient == nil {
		err = errors.New("yituClient not initialized")
		return
	}
	xl := xlog.FromContextSafe(ctx)
	decodedURI, err := gutil.LoadImage(snapshot.FeatureURI)
	if err != nil {
		xl.Errorf("yituFaceDetect.LoadImage: %+v", err)
		return
	}
	pts, err := h.yituClient.DetectFace(ctx, decodedURI)
	if err != nil {
		xl.Errorf("yituFaceDetect.DetectFace: %+v", err)
		return
	}
	encodedRawImage = []byte(base64.StdEncoding.EncodeToString(decodedURI))
	for _, pt := range pts {
		facePts := pt.ToPTS()
		faces = append(faces, proto.DeeperFacePeopleInfo{
			PTS:     facePts,
			FaceUri: encodedRawImage,
		})
	}

	return
}

// 人脸搜索
func (h *FaceSearchHandler) faceSearch(ctx context.Context, imageData []byte, pts [4][2]int) (output []proto.YituSearchResultItem, err error) {
	if h.yituClient == nil {
		err = errors.New("yituClient not initialized")
		return
	}
	return h.yituClient.SearchFace(ctx, imageData, pts)
}

// 截取图片
func clipImage(decodedURI []byte, pts [4][2]int) (encodedURI []byte, err error) {
	m, fm, err := image.Decode(bytes.NewBuffer(decodedURI))
	if err != nil {
		return nil, err
	}

	emptyBuff := bytes.NewBuffer(make([]byte, 0)) //开辟一个新的空buff
	switch fm {
	case "jpeg", "jpg":
		rgbImg := m.(*image.YCbCr)
		subImg := rgbImg.SubImage(image.Rect(pts[0][0], pts[0][1], pts[2][0], pts[2][1])).(*image.YCbCr)
		jpeg.Encode(emptyBuff, subImg, nil) //img写入到buff
	default:
		err := errors.New("图片格式错误")
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(emptyBuff.Bytes())), nil
}
