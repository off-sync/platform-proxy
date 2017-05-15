package setcertificate

type Command interface {
	Execute(model *CommandModel) error
}
