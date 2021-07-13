package puzzle

type DefaultRender struct {
    ErrNo  int         `json:"errNo"`
    ErrMsg string      `json:"errStr"`
    Data   interface{} `json:"data"`
}
