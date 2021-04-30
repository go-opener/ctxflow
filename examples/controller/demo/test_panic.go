package demo

import (
    //"github.com/go-opener/ctxflow/v2"
    "fmt"
    "github.com/go-opener/ctxflow/v2"
    "github.com/go-opener/ctxflow/v2/layer"
    //"fmt"
)

type TestPanic struct {
    layer.Controller
}

func (entity *TestPanic) Action() {

    fmt.Printf("%+v",string(ctxflow.StackTrace()))
    entity.RenderJsonSucc("sss")
    //panic("wrong!")
}
