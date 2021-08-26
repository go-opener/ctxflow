package puzzle

type DefaultRender struct {
    ErrNo  int         `json:"errNo"`
    ErrMsg string      `json:"errMsg"`
    Data   interface{} `json:"data"`
}
