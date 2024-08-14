package loafergo_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go/v2"
)

func TestSQSError_Error(t *testing.T) {
	t.Run("Package error Error", func(t *testing.T) {
		err := loafergo.ErrGetMessage
		assert.Equal(t, "unable to retrieve message", err.Error())
	})

	t.Run("Package error Context", func(t *testing.T) {
		errPkg := loafergo.ErrGetMessage
		err := errPkg.Context(fmt.Errorf("some error"))
		assert.NotNil(t, err)
		assert.Equal(t, "unable to retrieve message: some error", err.Error())
	})
}
