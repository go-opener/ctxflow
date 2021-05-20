package dsTestLog

import "github.com/go-opener/ctxflow/v2/layer"

type TestLogRepository struct {
    layer.DataSet
}

func (entity *TestLogRepository) DebugFunction() string {
    entity.LogDebug("this is domin log")

    return "ok"
}

