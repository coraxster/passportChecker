package passportChecker

type ExistChecker interface {
	Add([]string) error
	Check([]string) ([]bool, error)
}
