package client

type IFileServer interface {
	Init() error
	Save(key string, data []byte) (url string, err error)
	// GetURI(key string) (string, error)
}

type MockFileServer struct{}

var _ IFileServer = &MockFileServer{}

func NewMockFileServer() *MockFileServer {
	return &MockFileServer{}
}

func (MockFileServer) Init() error {
	return nil
}

func (MockFileServer) Save(key string, data []byte) (url string, err error) {
	return "http://mock/url", nil
}
