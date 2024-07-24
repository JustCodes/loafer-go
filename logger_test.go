package loafergo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerFunc_Log(t *testing.T) {
	expectedHandlers := []string{"name1"}
	var loggedHandlers []string
	LoggerFunc(func(args ...interface{}) {
		loggedHandlers = append(loggedHandlers, args[0].(string))
	}).Log("name1")
	assert.Equal(t, expectedHandlers, loggedHandlers)
}

func TestDefaultLogger_Log(t *testing.T) {
	log := newDefaultLogger()
	log.Log("log")
	assert.Implements(t, (*Logger)(nil), log)
}
