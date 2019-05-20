package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"sync"
	"time"

	"qiniu.com/vas-app/biz/proto"

	"github.com/golang/freetype"
	"github.com/qiniu/log.v1"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/vas-app/biz/analyzer/client"
	"qiniu.com/vas-app/biz/analyzer/dao"
	"qiniu.com/vas-app/util"
)

type EventType int

type snapshot struct {
	body    *proto.ImageBody
	jpgData []byte
	pts     [][2]int
	class   int //非机动车类别
}

type nonMotorValue struct {
	id        int
	endTime   time.Time
	msgs      map[EventType]proto.TrafficEventMsg //可能同一辆车多种违法
	startTime time.Time
	count     int
	// lastPts    [][2]int
	bestImage  snapshot //biggest image body
	firstImage snapshot
	lastImage  snapshot
}

type NonMotorHandlerConfig struct {
	Fileserver string `json:"fileserver"`
	Timeout    int    `json:"timeout"` //超时时间，超过多少秒没有同一id的人认为结束
}

//算法区域人员总数
type NonMotorHandler struct {
	xlog     *log.Logger
	task     *proto.Task
	m        map[int]*nonMotorValue //外卖人员
	config   *NonMotorHandlerConfig
	cancel   context.CancelFunc
	mutex    sync.RWMutex
	eventDao dao.EventDao
	fs       client.IFileServer
}

func NewNonMotorHandler(ctx context.Context, eventDao dao.EventDao, task *proto.Task, fileserver string) (handler *NonMotorHandler, err error) {
	xlog, ok := ctx.Value("xlog").(*log.Logger)
	if !ok {
		xlog = log.Std
		xlog.Error("Get context log error !")
	}
	config := &NonMotorHandlerConfig{
		Fileserver: fileserver,
		Timeout:    7, //todo
	}

	if err != nil {
		xlog.Errorf("failed to create dao:%v", err)
		return nil, err
	}
	fs, err := client.NewFileserver(fileserver)
	if err != nil {
		xlog.Errorf("failed to create fileserver instance:%v", err)
		return nil, err
	}
	handler = &NonMotorHandler{
		xlog:     xlog,
		task:     task,
		config:   config,
		m:        make(map[int]*nonMotorValue),
		mutex:    sync.RWMutex{},
		eventDao: eventDao,
		fs:       fs,
	}

	ctx1, cancel := context.WithCancel(ctx)
	handler.cancel = cancel
	handler.init(ctx1)
	return handler, nil
}

func (h *NonMotorHandler) CanHandle(modelKey string) bool {
	return modelKey == NonMotorHandlerModelKey
}

func (h *NonMotorHandler) startMsgSaver(ctx context.Context, saverCh chan nonMotorValue) {
	xlog := h.xlog
	//save msg and image data
	for i := 0; i < 10; i++ {

		go func(routineNum int) {
		Loop:
			for {
				select {
				case <-ctx.Done():
					xlog.Println("stop nonmotor handler event go routine :", routineNum)
					break Loop
				case msg := <-saverCh:
					{
						xlog.Println("recieved msg:", msg.msgs)
						uid := util.GenUuid()
						var err error
						{
							err = h.parseImage(&msg.bestImage)
							if err != nil {
								xlog.Println("parse best image data error:", err)
								continue
							}
							err = h.parseImage(&msg.firstImage)
							if err != nil {
								xlog.Println("parse first image data error:", err)
								continue
							}
							err = h.parseImage(&msg.lastImage)
							if err != nil {
								xlog.Println("parse last image data error:", err)
								continue
							}
						}

						imgUrlRaw, err := h.fs.Save(uid+"_raw.jpg", msg.bestImage.jpgData)
						if err != nil {
							xlog.Println("upload raw image  error:", err)
							continue
						}
						firstImgUrlRaw, err := h.fs.Save(uid+"_first_raw.jpg", msg.firstImage.jpgData)
						if err != nil {
							xlog.Println("upload  first snapshot raw image error:", err)
						}
						lastImgUrlRaw, err := h.fs.Save(uid+"_last_raw.jpg", msg.lastImage.jpgData)
						if err != nil {
							xlog.Println("upload last snapshot raw image error:", err)
						}

						//extract image
						sizeRatio, ok := h.task.Config["upload_image_size_ratio"].(float64)
						if !ok {
							sizeRatio = 1.0
						}
						rects := []int{}
						for _, i := range msg.bestImage.pts {
							for _, j := range i {
								rects = append(rects, int(sizeRatio*float64(j)))
							}
						}
						featureImg, err := util.ExtractImage(msg.bestImage.jpgData, rects)
						if err != nil {
							xlog.Println("extract image error:", err)
							continue
						}

						featureUrl, err := h.fs.Save(uid+"_feature.jpg", featureImg)
						if err != nil {
							xlog.Println("upload feature image error:", err)
							continue
						}

						// xlog.Println(featureUrl, msg.id)
						for _, v := range msg.msgs {
							xlog.Printf("msg:%+v\n", v)
							iUid := fmt.Sprintf("%s_%d", uid, v.EventType)

							//画框和违法行为标记
							body, err := drawDetectInfo(msg.bestImage.jpgData, msg.bestImage.pts, int(v.EventType), msg.bestImage.class)
							if err != nil {
								xlog.Println("draw image tag info error:", err)
								continue
							}
							url, err := h.fs.Save(iUid+".jpg", body)
							if err != nil {
								xlog.Println("upload image error:", err)
								continue
							}

							body, err = drawDetectInfo(msg.firstImage.jpgData, msg.firstImage.pts, int(v.EventType), msg.firstImage.class)
							if err != nil {
								xlog.Println("draw image tag info error:", err)
								continue
							}

							firstImageUrl, err := h.fs.Save(iUid+"_first.jpg", body)
							if err != nil {
								xlog.Println("upload first snapshot image error:", err)
								continue
							}

							body, err = drawDetectInfo(msg.lastImage.jpgData, msg.lastImage.pts, int(v.EventType), msg.lastImage.class)
							if err != nil {
								xlog.Println("draw image tag info error:", err)
								continue
							}
							lastImageUrl, err := h.fs.Save(iUid+"_last.jpg", body)
							if err != nil {
								xlog.Println("upload last snapshot image error:", err)
								continue
							}
							v.EventID = fmt.Sprintf("%s-%d-%d", uid, msg.id, v.EventType)
							v.Snapshot = []proto.Snapshot{
								proto.Snapshot{
									SnapshotURI:    url,
									SnapshotURIRaw: imgUrlRaw,
									FeatureURI:     featureUrl,
									Pts:            msg.bestImage.pts,
									Class:          msg.bestImage.class,
								},
								proto.Snapshot{
									SnapshotURI:    firstImageUrl,
									SnapshotURIRaw: firstImgUrlRaw,
									Pts:            msg.firstImage.pts,
								},
								proto.Snapshot{
									SnapshotURI:    lastImageUrl,
									SnapshotURIRaw: lastImgUrlRaw,
									Pts:            msg.lastImage.pts,
								},
							}
							v.StartTime = msg.startTime
							v.EndTime = msg.endTime
							v.Status = proto.StatusInit
							v.Mark = proto.Mark{
								Marking: proto.MarkingInit,
							}

							err = h.eventDao.Insert(v)
							if err != nil {
								xlog.Println("insert msg to db error:", err)
								continue
							}
							xlog.Println("insert db:", v)
						}
					}
				}
			}
			xlog.Println("stop save msg and image data go routine: ", routineNum)
		}(i)
	}

}

func (h *NonMotorHandler) init(ctx context.Context) {
	xlog := h.xlog
	var saverCh = make(chan nonMotorValue, 20)

	h.startMsgSaver(ctx, saverCh)

	go func() {
	Loop:
		for {
			select {
			case <-ctx.Done():
				xlog.Println("stop nonmotor handler event go routine ")
				break Loop
			default:
				now := time.Now()
				removeList := []int{}

				h.mutex.RLock()
				for k, v := range h.m {
					if v == nil {
						removeList = append(removeList, k)
						continue
					}
					if now.Sub(v.endTime) > time.Second*time.Duration(h.config.Timeout) {
						xlog.Println("send to msg saver chan:", k, v.msgs)
						saverCh <- *v
						removeList = append(removeList, k)
						v = nil
					}
				}
				h.mutex.RUnlock()

				if len(removeList) != 0 {
					h.mutex.Lock()
					xlog.Println("removing map keys :", removeList)
					for _, k := range removeList {
						delete(h.m, k)
					}
					h.mutex.Unlock()
				}
			}
			time.Sleep(time.Second * 10)
		}
	}()
}

func (h *NonMotorHandler) Handle(data interface{}, body *proto.ImageBody) (err error) {
	xlog := h.xlog
	if body == nil {
		xlog.Println("empty body ,skip")
		return nil
	}
	bData, err := json.Marshal(data)
	if err != nil {
		xlog.Println(err)
		return err
	}
	var wmData proto.NonmotorModelData
	err = json.Unmarshal(bData, &wmData)
	if err != nil {
		xlog.Println(err)
		return err
	}
	xlog.Println(wmData)

	now := time.Now()

	h.mutex.Lock()

	for _, v := range wmData.Boxes {
		xlog.Println(v)
		if info, ok := h.m[v.ID]; ok && info != nil {
			info.count++
			info.endTime = now
			//业务方需要连续的图片，此处暂时设为第一张图片后的第5张
			if info.count < 5 {
				info.lastImage = snapshot{
					body:  body,
					pts:   v.Pts,
					class: v.Class,
				}
			}

			if b, err := h.isBiggerRect(v.Pts, info.bestImage.pts); b && err == nil {
				xlog.Println("bigger rect ,replace old data")
				h.parseData(&v, info.msgs)
				info.bestImage.pts = v.Pts
				info.bestImage.body = body
				info.bestImage.class = v.Class
			}
		} else {
			snap := snapshot{
				body:  body,
				pts:   v.Pts,
				class: v.Class,
			}
			h.m[v.ID] = &nonMotorValue{
				id:         v.ID,
				startTime:  now,
				endTime:    now,
				firstImage: snap,
				lastImage:  snap,
				bestImage:  snap,
				msgs:       map[EventType]proto.TrafficEventMsg{},
			}
			h.parseData(&v, h.m[v.ID].msgs)
		}
	}
	h.mutex.Unlock()
	return nil
}

func (h *NonMotorHandler) parseData(input *proto.NonmotorModelBox, output map[EventType]proto.TrafficEventMsg) {
	for _, violation := range input.Violation {
		log.Println(violation)
		msg := proto.TrafficEventMsg{
			Region:    h.task.Region, // 补充区域信息
			EventType: violation,
			CameraID:  h.task.CameraID,
			Address:   h.task.StreamAddr,
		}
		output[EventType(msg.EventType)] = msg
	}
}

func (h *NonMotorHandler) isBiggerRect(a, b [][2]int) (bool, error) {
	if len(a) != 2 || len(b) != 2 {
		return false, errors.New("invalid input")
	}
	return abs((a[1][0]-a[0][0])*(a[1][1]*a[0][1])) > abs((b[1][0]-b[0][0])*(b[1][1]*b[0][1])), nil
}

func (h *NonMotorHandler) Release() error {
	h.cancel()
	h.xlog.Println("waimai handler release called")
	return nil
}

func (h *NonMotorHandler) parseImage(snap *snapshot) error {
	img, err := ConvertGBRData2Image(snap.body.Body, snap.body.Width, snap.body.Height)
	if err != nil {
		xlog.Error("convert image error:", err)
		return err
	}
	body, err := GetJpgData(img)
	if err != nil {
		xlog.Error("get jpg image data error:", err)
		return err
	}
	snap.jpgData = body
	return nil
}

type drawDot func(x, y int)

func drawDetectInfo(imgData []byte, points [][2]int, eventType int, class int) (ret []byte, err error) {
	if len(points) != 2 {
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
	clr := proto.MapTrafficDetectClassColor(class)
	drawD := func(x, y int) {
		for i := x - 1; i <= x+1; i++ {
			for j := y - 1; j <= y+1; j++ {
				lineImg.Set(i, j, clr)
			}
		}
	}
	drawLine(points[0][0], points[0][1], points[1][0], points[0][1], drawD)
	drawLine(points[1][0], points[0][1], points[1][0], points[1][1], drawD)
	drawLine(points[1][0], points[1][1], points[0][0], points[1][1], drawD)
	drawLine(points[0][0], points[1][1], points[0][0], points[0][1], drawD)

	c := image.NewRGBA(img.Bounds())
	draw.Draw(c, img.Bounds(), img, image.ZP, draw.Src)
	draw.Draw(c, img.Bounds(), lineImg, image.ZP, draw.Over)

	ft, err := util.GetFont("")
	if err != nil {
		log.Println("get font error:", err)
		return nil, err
	}
	fc := freetype.NewContext()
	// c.SetDPI(72)
	fc.SetFont(ft)
	fc.SetFontSize(40)
	fc.SetClip(img.Bounds())
	fc.SetDst(c)
	fc.SetSrc(image.NewUniform(clr))

	pt := freetype.Pt(points[1][0]+5, points[0][1])
	_, err = fc.DrawString(proto.MapEventType(eventType), pt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	pt = freetype.Pt(points[1][0]+5, points[1][1])
	_, err = fc.DrawString(proto.MapTrafficDetectClass(class), pt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

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
