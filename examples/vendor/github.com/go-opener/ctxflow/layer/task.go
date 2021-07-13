package layer

type ITask interface {
    IFlow
    Run(args []string) error
}

type Task struct {
    Flow
}

func (entity *Task) Run(args []string) error{
    return nil
}

