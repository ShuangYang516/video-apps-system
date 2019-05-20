package proto

import "fmt"

type YituLoginResp struct {
	ResultCode    string `json:"resultCode" bson:"resultCode"`
	ResultMessage string `json:"resultMessage" bson:"resultMessage"`
}

type YituLibListReq struct {
	UserID string `json:"userId" bson:"userId"` // 固定传值 20
}

type YituLibItem struct {
	ID           int    `json:"id" bson:"id"`
	YituID       int    `json:"yituId" bson:"yituId"`
	Name         string `json:"name" bson:"name"`
	Type         string `json:"type" bson:"type"`
	Desc         string `json:"desc" bson:"desc"`
	CreateTime   string `json:"createTime" bson:"createTime"`
	FaceImageNum int    `json:"faceImageNum" bson:"faceImageNum"`
}

type YituLibListResp struct {
	ResultCode    string        `json:"resultCode" bson:"resultCode"`
	ResultMessage string        `json:"resultMessage" bson:"resultMessage"`
	Data          []YituLibItem `json:"data" bson:"data"`
}

type YituFaceRect struct {
	W int `json:"w" bson:"w"`
	H int `json:"h" bson:"h"`
	X int `json:"x" bson:"x"`
	Y int `json:"y" bson:"y"`
}

func (rect YituFaceRect) FromPTS(pts [4][2]int) (YituFaceRect, bool) {
	rect.X = pts[0][0]
	rect.Y = pts[0][1]
	rect.W = pts[2][0] - pts[0][0]
	rect.H = pts[2][1] - pts[0][1]
	return rect, rect.W > 0 && rect.H > 0
}

func (rect YituFaceRect) String() string {
	return fmt.Sprintf(`{"w":%d,"h":%d,"x":%d,"y":%d}`, rect.W, rect.H, rect.X, rect.Y)
}

func (rect YituFaceRect) ToPTS() (pts [4][2]int) {
	pts[0][0] = rect.X
	pts[0][1] = rect.Y
	pts[1][0] = rect.X + rect.W
	pts[1][1] = rect.Y
	pts[2][0] = rect.X + rect.W
	pts[2][1] = rect.Y + rect.H
	pts[3][0] = rect.X
	pts[3][1] = rect.Y + rect.H
	return
}

type YituFeatureCheckFaceItem struct {
	FaceInfo []YituFaceRect `json:"faceInfo" bson:"faceInfo"`
	Image    string         `json:"image" bson:"image"`
}

type YituFeatureCheckFaceResp struct {
	ResultCode    string                   `json:"resultCode" bson:"resultCode"`
	ResultMessage string                   `json:"resultMessage" bson:"resultMessage"`
	Data          YituFeatureCheckFaceItem `json:"data" bson:"data"`
}

type YituSearchReq struct { // NOTE 这个请求时form形式
	YituIds   []string `json:"yituIds" bson:"yituIds"`
	ImageData string   `json:"imageData" bson:"imageData"`
}

type YituSearchResultItem struct {
	YituID           string  `json:"yituId" bson:"yituId"`
	YituFaceImageUri string  `json:"yituFaceImageUri" bson:"yituFaceImageUri"`
	Name             string  `json:"name" bson:"name"`
	Nation           string  `json:"nation" bson:"nation"`
	Sex              string  `json:"sex" bson:"sex"` //  1: Male, 0: Female
	IDCard           string  `json:"idCard" bson:"idCard"`
	Similarity       float64 `json:"similarity" bson:"similarity"`
	Offset           int     `json:"offset" bson:"offset"`

	FaceImageId  string `json:"faceImageId" bson:"faceImageId"`
	FaceImageUrl string `json:"faceImageUrl" bson:"faceImageUrl"`
	FaceRect     struct {
		W int `json:"w" bson:"w"`
		H int `json:"h" bson:"h"`
		X int `json:"x" bson:"x"`
		Y int `json:"y" bson:"y"`
	} `json:"faceRect" bson:"faceRect"`

	FaceLibId   int    `json:"faceLibId" bson:"faceLibId"`
	FaceLibName string `json:"faceLibName" bson:"faceLibName"`

	UpdateTime string `json:"updateTime" bson:"updateTime"`
}

type YituSearchResult struct {
	RetrievalQueryId string                 `json:"retrievalQueryId" bson:"retrievalQueryId"`
	TotalRow         int                    `json:"totalRow" bson:"totalRow"`
	PageNumber       int                    `json:"pageNumber" bson:"pageNumber"`
	TotalPage        int                    `json:"totalPage" bson:"totalPage"`
	PageSize         int                    `json:"pageSize" bson:"pageSize"`
	List             []YituSearchResultItem `json:"list" bson:"list"`
}

type YituSearchResp struct {
	ResultCode    string           `json:"resultCode" bson:"resultCode"`
	ResultMessage string           `json:"resultMessage" bson:"resultMessage"`
	Data          YituSearchResult `json:"data" bson:"data"`
}
