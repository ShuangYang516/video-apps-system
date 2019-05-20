package client

import (
	"log"

	simplejson "github.com/bitly/go-simplejson"
	"qiniu.com/vas-app/util"
)

type FileServer struct {
	host string
}

var _ IFileServer = &FileServer{host: ""}

func NewFileserver(host string) (*FileServer, error) {
	server := &FileServer{
		host: host,
	}
	return server, nil
}

func (s *FileServer) Init() error {
	return nil
}

func (s *FileServer) Save(filename string, body []byte) (string, error) {
	return s.UploadStream(body, filename)
}

// //TODO
// func (s *FileServer) GetURI(key string) (string, error) {
// 	return "", fmt.Errorf("not implemented")
// }

func (s *FileServer) UploadStream(body []byte, filename string) (string, error) {
	header := map[string]string{
		"filename": filename,
	}
	resp, err := util.PostRaw(s.host+"/v1/upload/stream", body, header)
	if err != nil {
		log.Println(err)
		return "", err
	}
	log.Println(string(resp))
	sj, err := simplejson.NewJson(resp)
	if err != nil {
		log.Println(err)
		return "", err
	}

	url, err := sj.Get("extra").Get("uri").String()
	if err != nil {
		log.Println(err)
		return "", err
	}

	return url, nil
}
