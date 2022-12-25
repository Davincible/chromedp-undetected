package easyjson

import (
	"testing"

	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/require"
)

func TestEasyJSONMap(t *testing.T) {
	a := map[string]any{
		"one": 1.8,
		"two": "two",
	}

	b := New(a)

	j, err := easyjson.Marshal(b)
	require.NoError(t, err, "marshal")
	t.Log(string(j))

	n := New(map[string]any{})

	require.NoError(t, easyjson.Unmarshal(j, n))
	require.Equal(t, a, n.Data)
}

func TestEasyJSONSlice(t *testing.T) {
	a := []string{"a", "b", "c", "d"}

	b := New(a)

	j, err := easyjson.Marshal(b)
	require.NoError(t, err, "marshal")
	t.Log(string(j))

	n := New([]string{})

	require.NoError(t, easyjson.Unmarshal(j, n))
	require.Equal(t, a, n.Data)
}
