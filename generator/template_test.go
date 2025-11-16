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

		{useProtoNames: true, input: "k8s_field", want: "k8s_field"},
		{useProtoNames: true, input: "foo_bar", want: "foo_bar"},
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

func TestEscapeJSDoc(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single wildcard slash",
			input: "/api/v1/{name=customers/*/secrets}",
			want:  "/api/v1/{name=customers/*\\/secrets}",
		},
		{
			name:  "multiple wildcard slashes",
			input: "/api/v1/{name=customers/*/profiles/*/secrets}",
			want:  "/api/v1/{name=customers/*\\/profiles/*\\/secrets}",
		},
		{
			name:  "triple wildcard slashes",
			input: "/api/v2/{name=a/*/b/*/c/*/items}",
			want:  "/api/v2/{name=a/*\\/b/*\\/c/*\\/items}",
		},
		{
			name:  "no wildcard slashes",
			input: "/api/v1/users",
			want:  "/api/v1/users",
		},
		{
			name:  "trailing wildcard without slash",
			input: "/api/v1/{name=users/*}",
			want:  "/api/v1/{name=users/*}",
		},
		{
			name:  "comment closer in text",
			input: "GET /api/*/",
			want:  "GET /api/*\\/",
		},
		{
			name:  "no escaping needed",
			input: "/api/v1/resource",
			want:  "/api/v1/resource",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeJSDoc(tt.input)
			assert.Equal(t, tt.want, got, "escapeJSDoc(%s) = %s, want %s", tt.input, got, tt.want)
		})
	}
}
