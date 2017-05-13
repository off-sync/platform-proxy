package getfrontends

type Query interface {
	Execute(model *QueryModel) (*ResultModel, error)
}
