package command

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Kubectl(args ...string) *exec.Cmd {
	return exec.Command("kubectl", args...)
}

func MustRun(t *testing.T, cmd *exec.Cmd) string {
	out, err := runWithOutput(cmd)
	require.NoErrorf(t, err, "Running command %q.", getCommandDescription(cmd))
	return out
}

func getCommandDescription(cmd *exec.Cmd) string {
	basename := path.Base(cmd.Path)
	args := strings.Join(cmd.Args, " ")
	return fmt.Sprintf("%s %s", basename, args)
}

// runWithOutput executes the command returning its standard output.
func runWithOutput(cmd *exec.Cmd) (string, error) {
	b := new(bytes.Buffer)
	cmd.Stdout = b
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
