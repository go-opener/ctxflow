package demo

import (
    //"github.com/go-opener/ctxflow"
    "fmt"
    "github.com/go-opener/ctxflow"
    "github.com/go-opener/ctxflow/layer"
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
