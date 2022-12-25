// Package easyjson provides utilities for use with easyjson.
package easyjson

import (
	"encoding/json"

	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

var _ easyjson.MarshalerUnmarshaler = (*GenericEJ[any])(nil)

// New creates a new geneic EasyJSON container.
func New[T any](data T) *GenericEJ[T] {
	return &GenericEJ[T]{data}
}

// GenericEJ is a type that uses the stdlib json (un)marshaler to make any
// type compatible with easyjson by implementing the EasyJSON interfaces.
type GenericEJ[T any] struct {
	Data T
}

// MarshalEasyJSON marshals any type to json with use of the stdlib json pkg.
func (t *GenericEJ[T]) MarshalEasyJSON(w *jwriter.Writer) {
	b, err := json.Marshal(t.Data)

	w.Buffer.AppendBytes(b)
	w.Error = err
}

// UnmarshalEasyJSON unmarshals any type from json with use of the stdlib json pkg.
func (t *GenericEJ[T]) UnmarshalEasyJSON(w *jlexer.Lexer) {
	if err := json.Unmarshal(w.Data, &t.Data); err != nil {
		w.AddError(err)
	}
}
