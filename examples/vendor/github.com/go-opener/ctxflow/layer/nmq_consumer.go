package layer

type INMQConsumer interface {
    IFlow
    Process() (interface{}, error)
}

type NmqConsumer struct {
    Flow
}
