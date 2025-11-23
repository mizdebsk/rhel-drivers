package api

//go:generate mockgen -source=exec.go -destination=../mocks/exec_mock.go -package=mocks

type Executor interface {
	Run(command string, args []string) error
	RunCapture(command string, args ...string) ([]string, error)
}
