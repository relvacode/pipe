package console

type Expression interface {
	Eval(interface{}) (interface{}, error)
}

type DefaultExpression struct {
}

func (DefaultExpression) Eval(x interface{}) (interface{}, error) {
	return x, nil
}
