package task

type Action func()
type ActionWith func(interface{})
type ErrAction func() error
type ErrActionWith func(interface{}) error
type Func func() interface{}
type FuncWith func(interface{}) interface{}
type ErrFunc func() (interface{}, error)
type ErrFuncWith func(interface{}) (interface{}, error)
