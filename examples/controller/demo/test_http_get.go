package demo

import (
    "examples/api/apiBaidu"
    "github.com/go-opener/ctxflow/layer"
)

type TestHttpGet struct {
    layer.Controller
}

func (entity *TestHttpGet) Action() {
    entity.LogInfo("test TestHttpGet start")

    baiduApi := entity.Use(new(apiBaidu.Baidu)).(*apiBaidu.Baidu)
    data,err := baiduApi.GetLiveStat()
    if err != nil {
        entity.RenderJsonFail(err)
        return
    }
    entity.RenderJsonSucc(data)
}
