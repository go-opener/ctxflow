package puzzle

type DefaultRender struct {
    ErrNo  int         `json:"errNo"`
    ErrMsg string      `json:"errStr"`
    Data   interface{} `json:"data"`
}

type IHTTPClient interface {
    Request(method string, uri string, reqBody []byte, header map[string]string, cookies map[string]string, contentType string) (data []byte, err error)
}