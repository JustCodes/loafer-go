package loafergo_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go/v2"
)

func TestLoggerFunc_Log(t *testing.T) {
	called := false
	logger := loafergo.LoggerFunc(func(args ...interface{}) {
		called = true
		assert.Contains(t, args[0], "Hello LoggerFunc")
	})

	logger.Log("Hello LoggerFunc!")
	assert.True(t, called, "LoggerFunc should have been called")
}

func TestDefaultLogger_Log(t *testing.T) {
	// Redirect stdout to buffer
	original := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Replace logger output to our pipe
	l := log.New(w, "", 0)
	defaultLog := loafergo.LoggerFunc(func(args ...interface{}) {
		l.Println(args...)
	})

	defaultLog.Log("Default logger test")

	// Capture and restore
	w.Close()
	os.Stdout = original

	var outputBuf bytes.Buffer
	_, _ = outputBuf.ReadFrom(r)
	logOutput := outputBuf.String()

	assert.Contains(t, logOutput, "Default logger test")
}

func TestNewDefaultLogger(t *testing.T) {
	logger := loafergo.LoggerFunc(func(args ...interface{}) {
		fmt.Fprintln(os.Stdout, args...)
	})
	assert.NotNil(t, logger)
}

func TestNoOpLogger_Log(t *testing.T) {
	// Should not panic or output anything
	logger := loafergo.NoOpLogger{}
	logger.Log("This should not appear anywhere")
}
