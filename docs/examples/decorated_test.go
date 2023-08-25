package examples

import (
    "errors"
    "fmt"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
)

func TestFailureWithoutDecoration(t *testing.T) {
    require.NoError(t, A(2))
    require.NoError(t, B(3))
    require.NoError(t, C(2))
}

func TestFailureWithDecoration(t *testing.T) {
    t.Run(fmt.Sprintf("Running A with sleep for %v", 2), func(t *testing.T) {
        require.NoError(t, A(2))
    })

    t.Run(fmt.Sprintf("Running B with sleep for %v", 3), func(t *testing.T) {
        require.NoError(t, B(3))
    })

    t.Run(fmt.Sprintf("Running C with sleep for %v", 2), func(t *testing.T) {
        require.NoError(t, C(2))
    })
}

func A(sleepSeconds time.Duration) error {
    time.Sleep(sleepSeconds * time.Second)
    return nil
}

func B(sleepSeconds time.Duration) error {
    time.Sleep(sleepSeconds * time.Second)
    return nil
}

func C(sleepSeconds time.Duration) error {
    time.Sleep(sleepSeconds * time.Second)
    return errors.New("failed")
}
