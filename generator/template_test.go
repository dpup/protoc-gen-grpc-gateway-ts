package generator

import (
	"testing"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/registry"
	"github.com/stretchr/testify/assert"
)

func TestFieldName(t *testing.T) {
	tests := []struct {
		useProtoNames bool
		input         string
		want          string
	}{
		{useProtoNames: false, input: "k8s_field", want: "k8sField"},

		{useProtoNames: false, input: "foo_bar", want: "fooBar"},
		{useProtoNames: false, input: "foobar", want: "foobar"},
		{useProtoNames: false, input: "foo_bar_baz", want: "fooBarBaz"},

		{useProtoNames: false, input: "foobar3", want: "foobar3"},
		{useProtoNames: false, input: "foo3bar", want: "foo3bar"},
		{useProtoNames: false, input: "foo3_bar", want: "foo3Bar"},
		{useProtoNames: false, input: "foo_3bar", want: "foo3bar"},
		{useProtoNames: false, input: "foo_3_bar", want: "foo3Bar"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := &registry.Registry{Options: registry.Options{UseProtoNames: tt.useProtoNames}}
			fn := fieldName(r)
			if got := fn(tt.input); got != tt.want {
				assert.Equal(t, tt.want, got, "fieldName(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}
