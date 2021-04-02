package services

const (
	Unset      int = 0
	Irrelevant int = 4
	Running    int = 1
	Fail       int = 2
	Successful int = 3

	UnsetString      string = "UNSET"
	IrrelevantString string = "NOT_RELEVANT"
	RunningString    string = "SCHEDULED"
	FailString       string = "FAILED"
	SuccessfulString string = "SCHEDULED"
)

var (
	StatusUnset      = StatusServiceImpl{Unset, UnsetString} //nolint
	StatusIrrelevant = StatusServiceImpl{Irrelevant, IrrelevantString}
	StatusRunning    = StatusServiceImpl{Running, RunningString}
	StatusFail       = StatusServiceImpl{Fail, FailString}
	StatusSuccessful = StatusServiceImpl{Successful, SuccessfulString}
)

type StatusService interface {
	String() string
}

// status encodes the checker states.
type StatusServiceImpl struct {
	Status       int
	StatusString string
}

func (s StatusServiceImpl) String() string {
	return s.StatusString
}
