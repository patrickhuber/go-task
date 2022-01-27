package task

type Action func()
type ActionWith func(interface{})
type Func func() interface{}
type FuncWith func(interface{}) interface{}
type ErrFunc func() (interface{}, error)
type ErrFuncWith func(interface{}) (interface{}, error)
