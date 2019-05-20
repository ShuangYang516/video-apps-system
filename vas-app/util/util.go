package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net"
)

// JsonStr
func JsonStr(obj interface{}) string {
	raw, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(raw)
}

// ConvByJson
func ConvByJson(src interface{}, dest interface{}) error {
	tmpbs, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(tmpbs, dest)
}

// GetLocalIP
func GetLocalIP() string {
	addrArr, err := net.InterfaceAddrs()
	if nil != err {
		return ""
	}
	for _, addr := range addrArr {
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

func Md5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
