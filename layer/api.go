package layer

import (
    "github.com/go-opener/ctxflow/puzzle"
)



type IApi interface {
    IFlow
}

type Api struct {
    Flow
    http puzzle.IHTTPClient
}

func (entity *Api) PreUse(args ...interface{}) {
    entity.Flow.PreUse(args...)
}





