package generator

import (
	"testing"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/data"
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

func TestWrapBytesFieldsInURL(t *testing.T) {
	tests := []struct {
		name          string
		useProtoNames bool
		url           string
		httpMethod    string
		inputType     string
		messages      []*data.Message
		want          string
	}{
		{
			name:          "single bytes field in URL path",
			useProtoNames: false,
			url:           `/api/v1/{encoded_path}`,
			inputType:     "TestRequest",
			messages: []*data.Message{
				{
					Name: "TestRequest",
					Fields: []*data.Field{
						{
							Name: "encoded_path",
							Type: "bytes",
						},
					},
				},
			},
			want: `/api/v1/${req.encodedPath ? new TextDecoder().decode(req.encodedPath) : ''}`,
		},
		{
			name:          "bytes field with useProtoNames",
			useProtoNames: true,
			url:           `/api/v1/{encoded_path}`,
			inputType:     "TestRequest",
			messages: []*data.Message{
				{
					Name: "TestRequest",
					Fields: []*data.Field{
						{
							Name: "encoded_path",
							Type: "bytes",
						},
					},
				},
			},
			want: `/api/v1/${req.encoded_path ? new TextDecoder().decode(req.encoded_path) : ''}`,
		},
		{
			name:          "non-bytes field in URL path",
			useProtoNames: false,
			url:           `/api/v1/{parent}`,
			inputType:     "TestRequest",
			messages: []*data.Message{
				{
					Name: "TestRequest",
					Fields: []*data.Field{
						{
							Name: "parent",
							Type: "string",
						},
					},
				},
			},
			want: `/api/v1/${req.parent}`,
		},
		{
			name:          "mixed bytes and non-bytes fields",
			useProtoNames: false,
			url:           `/api/v1/{parent}/artefacts/{encoded_path}`,
			inputType:     "TestRequest",
			messages: []*data.Message{
				{
					Name: "TestRequest",
					Fields: []*data.Field{
						{
							Name: "parent",
							Type: "string",
						},
						{
							Name: "encoded_path",
							Type: "bytes",
						},
					},
				},
			},
			want: `/api/v1/${req.parent}/artefacts/${req.encodedPath ? new TextDecoder().decode(req.encodedPath) : ''}`,
		},
		{
			name:          "multiple bytes fields",
			useProtoNames: false,
			url:           `/api/v1/{first_path}/items/{second_path}`,
			inputType:     "TestRequest",
			messages: []*data.Message{
				{
					Name: "TestRequest",
					Fields: []*data.Field{
						{
							Name: "first_path",
							Type: "bytes",
						},
						{
							Name: "second_path",
							Type: "bytes",
						},
					},
				},
			},
			want: `/api/v1/${req.firstPath ? new TextDecoder().decode(req.firstPath) : ''}/items/${req.secondPath ? new TextDecoder().decode(req.secondPath) : ''}`,
		},
		{
			name:          "GET request with bytes field generates query params",
			useProtoNames: false,
			url:           `/api/v1/{encoded_path}`,
			inputType:     "TestRequest",
			httpMethod:    "GET",
			messages: []*data.Message{
				{
					Name: "TestRequest",
					Fields: []*data.Field{
						{
							Name: "encoded_path",
							Type: "bytes",
						},
					},
				},
			},
			want: `/api/v1/${req.encodedPath ? new TextDecoder().decode(req.encodedPath) : ''}?${fm.renderURLSearchParams(req, ["encodedPath"])}`,
		},
		{
			name:          "no path parameters",
			useProtoNames: false,
			url:           `/api/v1/users`,
			inputType:     "TestRequest",
			messages: []*data.Message{
				{
					Name: "TestRequest",
					Fields: []*data.Field{
						{
							Name: "name",
							Type: "string",
						},
					},
				},
			},
			want: `/api/v1/users`,
		},
		{
			name:          "input message not found",
			useProtoNames: false,
			url:           `/api/v1/{path}`,
			inputType:     "UnknownRequest",
			messages: []*data.Message{
				{
					Name: "TestRequest",
					Fields: []*data.Field{
						{
							Name: "path",
							Type: "bytes",
						},
					},
				},
			},
			want: `/api/v1/${req.path}`,
		},
		{
			name:          "fully qualified message name",
			useProtoNames: false,
			url:           `/api/v1/{encoded_path}`,
			inputType:     "TestRequest",
			messages: []*data.Message{
				{
					Name: "foocorp.bar.TestRequest",
					Fields: []*data.Field{
						{
							Name: "encoded_path",
							Type: "bytes",
						},
					},
				},
			},
			want: `/api/v1/${req.encodedPath ? new TextDecoder().decode(req.encodedPath) : ''}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registry.Registry{Options: registry.Options{UseProtoNames: tt.useProtoNames}}

			// Create a mock method
			method := data.Method{
				URL:        tt.url,
				HTTPMethod: tt.httpMethod,
				Input: &data.MethodArgument{
					Type: tt.inputType,
				},
			}

			fn := wrapBytesFieldsInURL(r)
			got := fn(method, tt.messages)
			assert.Equal(t, tt.want, got)
		})
	}
}
