package loafergo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go"
)

func TestDataType_String(t *testing.T) {
	d := loafergo.DataType("Number")
	assert.Equal(t, "Number", d.String())
}

func TestConfig_NewCustomAttribute(t *testing.T) {
	t.Run("With data type string", func(t *testing.T) {
		got := loafergo.Config{}
		want := []loafergo.CustomAttribute{{
			Title:    "title",
			DataType: "String",
			Value:    "my title",
		}}
		err := got.NewCustomAttribute(loafergo.DataTypeString, "title", "my title")
		assert.NoError(t, err)
		assert.Equal(t, want, got.Attributes)
	})

	t.Run("With data type string error", func(t *testing.T) {
		got := loafergo.Config{}
		err := got.NewCustomAttribute(loafergo.DataTypeString, "title", 1.6)
		assert.NotNil(t, err)
		assert.ErrorIs(t, loafergo.ErrMarshal, err)
	})

	t.Run("With data type number", func(t *testing.T) {
		got := loafergo.Config{}
		want := []loafergo.CustomAttribute{{
			Title:    "title",
			DataType: "Number",
			Value:    "1",
		}}
		err := got.NewCustomAttribute(loafergo.DataTypeNumber, "title", 1)
		assert.NoError(t, err)
		assert.Equal(t, want, got.Attributes)
	})

	t.Run("With data type number error", func(t *testing.T) {
		got := loafergo.Config{}
		err := got.NewCustomAttribute(loafergo.DataTypeNumber, "title", 1.6)
		assert.NotNil(t, err)
		assert.ErrorIs(t, loafergo.ErrMarshal, err)
	})
}
