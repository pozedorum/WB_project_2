package builtins

import (
	"io"
	"strings"

	"task15/internal/core"
)

type EchoUtil struct{}

func NewEchoUtil() *EchoUtil {
	return &EchoUtil{}
}

func (echu EchoUtil) Name() string {
	return "echo"
}

func (echu *EchoUtil) Execute(args []string, _ core.Environment, _ io.Reader, stdout io.Writer) error {
	_, err := io.WriteString(stdout, strings.Join(args, " ")+"\n")

	return err
}
