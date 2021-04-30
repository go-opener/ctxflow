package domTestLog

import "github.com/go-opener/ctxflow/v2/layer"

type TestLogDomain struct {
    layer.Domain
}

func (entity *TestLogDomain) DebugFunction() string {
    entity.LogDebug("this is domin log")

    return "ok"
}

