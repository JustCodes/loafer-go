package loafergo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go"
)

func TestDefaultLogger_Log(t *testing.T) {
	expectedHandlers := []string{"name1"}
	var loggedHandlers []string
	loafergo.LoggerFunc(func(args ...interface{}) {
		loggedHandlers = append(loggedHandlers, args[0].(string))
	}).Log("name1")
	assert.Equal(t, expectedHandlers, loggedHandlers)
}
