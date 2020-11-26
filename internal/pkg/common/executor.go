package common



type Executor interface {
	Prepare(job Job)	error
	Run() error
	Wait() error
	SendError(err error)
	Cleanup() error
}

type BuildError struct {
	Inner         error

}

func (b *BuildError) Error() string {
	if b.Inner == nil {
		return "error"
	}

	return b.Inner.Error()
}
type ExecutorProvider interface {
	CanCreate() bool
	Create() Executor
}

var executors map[string]ExecutorProvider


func GetExecutor(executorStr string) ExecutorProvider{
	if executors == nil {
		return nil
	}
	executor, _ := executors[executorStr]
	return executor
}


func RegisterExecutor(executor string, provider ExecutorProvider) {
	if executors == nil {
		executors = make(map[string]ExecutorProvider)
	}
	if _, ok := executors[executor]; ok {
		panic("Executor already exist: " + executor)
	}
	executors[executor] = provider
}
//func RegisterExecutor(executorStr string, executor Executor) {
//	logrus.Debugln("Registering", executor, "executor...")
//	if executors == nil {
//		executors = make(map[string]Executor)
//	}
//	if _, ok := executors[executorStr]; ok {
//		logrus.Panicln("Executor already exist: " + executorStr)
//	}
//	executors[executorStr] = executor
//}