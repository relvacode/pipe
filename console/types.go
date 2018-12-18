package console

type Expression interface {
	Eval(interface{}) (interface{}, error)
}
