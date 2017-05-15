package getservices

type Query interface {
	Execute(model *QueryModel) (*ResultModel, error)
}
