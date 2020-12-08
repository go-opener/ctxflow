package demo

import (
    //"fmt"
    "github.com/go-opener/ctxflow/layer"
)

type TestPanic struct {
    layer.Controller
}

func (entity *TestPanic) Action() {
    panic("wrong!")
}
