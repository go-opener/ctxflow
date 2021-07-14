package layer

type IApi interface {
    IFlow
}

type Api struct {
    Flow
}

func (entity *Api) PreUse(args ...interface{}) {
    entity.Flow.PreUse(args...)
}





