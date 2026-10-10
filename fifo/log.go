package fifo

import "time"

type LogReason uint

const (
	LogReasonInit LogReason = iota
	LogReasonResize
	LogReasonRead
	LogReasonWrite
)

type Log struct {
	Reason LogReason

	NewSize    *uint
	WriteValue *uint32
}

func (l *Log) String() string {

}
func (l *Log) Time() time.Time {
}
func (l *Log) ManualAction() bool {
}

func LogInit(size uint) *Log {
	return &Log{
		Reason:  LogReasonInit,
		NewSize: &size,
	}
}

func LogResize(size uint) *Log {
	return &Log{
		Reason:  LogReasonResize,
		NewSize: &size,
	}
}

func LogRead() *Log {
	return &Log{
		Reason: LogReasonRead,
	}
}

func LogWrite(value uint32) *Log {
	return &Log{
		Reason:     LogReasonWrite,
		WriteValue: &value,
	}
}
