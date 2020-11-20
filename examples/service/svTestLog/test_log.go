package svTestLog

import (
    "examples/domain/domTestLog"
    "github.com/go-opener/ctxflow/layer"
)

type TestLogService struct {
    layer.Service
}

func (entity *TestLogService) DebugFunction() string {
    entity.LogInfof("this is a Service log,service name:%+v","TestLogService")
    testLogDomain := entity.Use(new(domTestLog.TestLogDomain)).(*domTestLog.TestLogDomain)
    return testLogDomain.DebugFunction()
}

