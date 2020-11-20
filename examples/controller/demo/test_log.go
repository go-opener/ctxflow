package demo

import (
    //"fmt"
    "examples/service/svTestLog"
    "github.com/go-opener/ctxflow/layer"
)

type TestLog struct {
    layer.Controller
}

func (entity *TestLog) Action() {
    entity.LogWarn("this is a log")

    testLog := entity.Use(new(svTestLog.TestLogService)).(*svTestLog.TestLogService)

    entity.RenderJsonSucc(testLog.DebugFunction())
}
