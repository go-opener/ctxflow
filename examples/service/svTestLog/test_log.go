package svTestLog

import (
    "examples/data/dsTestLog"
    "github.com/go-opener/ctxflow/layer"
)

type TestLogService struct {
    layer.Service
}

func (entity *TestLogService) DebugFunction() string {
    entity.LogInfof("this is a Service log,service name:%+v","TestLogService")
    testLogRepo := entity.Use(new(dsTestLog.TestLogRepository)).(*dsTestLog.TestLogRepository)
    return testLogRepo.DebugFunction()
}

