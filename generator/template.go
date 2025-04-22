package generator

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/data"
	"github.com/dpup/protoc-gen-grpc-gateway-ts/registry"

	_ "embed"
)

//go:embed service.ts.tmpl
var serviceTmplScript string

//go:embed fetch_tmpl.ts
var fetchTmplScript string

const fetchTmplHeader = `{{- if not .EnableStylingCheck}}
/* eslint-disable */
// @ts-nocheck
{{- else -}}
/* eslint-disable @typescript-eslint/consistent-type-definitions */
{{- end}}
`

var fetchTmpl = fetchTmplHeader + fetchTmplScript

// Data object injected into the templates.
type TemplateData struct {
	*data.File
	EnableStylingCheck bool
	UseStaticClasses   bool
}

// ServiceTemplate gets the template for the primary typescript file.
func ServiceTemplate(r *registry.Registry) *template.Template {
	t := template.New("file")
	t = t.Funcs(sprig.TxtFuncMap())

	t = t.Funcs(template.FuncMap{
		"include": include(t),
		"tsType": func(fieldType data.Type) string {
			return tsType(r, fieldType)
		},
		"tsTypeKey":    tsTypeKey(r),
		"tsTypeDef":    tsTypeDef(r),
		"renderURL":    renderURL(r),
		"buildInitReq": buildInitReq,
		"fieldName":    fieldName(r),
		"functionCase": functionCase,
	})

	t = template.Must(t.Parse(serviceTmplScript))
	return t
}

// FetchModuleTemplate returns the go template for fetch module.
func FetchModuleTemplate() *template.Template {
	t := template.New("fetch")
	return template.Must(t.Parse(fetchTmpl))
}

func fieldName(r *registry.Registry) func(name string) string {
	return func(name string) string {
		if r.UseProtoNames {
			return name
		}
		return JSONCamelCase(name)
	}
}

var (
	// Match {field} or {field=pattern}, and return's param and pattern.
	pathParamRegexp = regexp.MustCompile(`{([^=}/]+)(?:=([^}]+))?}`)
)

func renderURL(r *registry.Registry) func(method data.Method) string {
	fieldNameFn := fieldName(r)
	return func(method data.Method) string {
		methodURL := method.URL
		matches := pathParamRegexp.FindAllStringSubmatch(methodURL, -1)
		fieldsInPath := make([]string, 0, len(matches))
		if len(matches) > 0 {
			slog.Debug("url matches", slog.Any("matches", matches))
			for _, m := range matches {
				expToReplace := m[0]
				fieldNameRaw := m[1]
				fieldName := fieldNameFn(fieldNameRaw)
				part := fmt.Sprintf(`${req.%s}`, fieldName)
				methodURL = strings.ReplaceAll(methodURL, expToReplace, part)
				fieldsInPath = append(fieldsInPath, fmt.Sprintf(`"%s"`, fieldName))
			}
		}
		urlPathParams := fmt.Sprintf("[%s]", strings.Join(fieldsInPath, ", "))

		if !method.ClientStreaming && (method.HTTPMethod == "GET" || method.HTTPMethod == "DELETE") {
			// parse the url to check for query string
			parsedURL, err := url.Parse(methodURL)
			if err != nil {
				return methodURL
			}
			renderURLSearchParamsFn := fmt.Sprintf("${fm.renderURLSearchParams(req, %s)}", urlPathParams)
			// prepend "&" if query string is present otherwise prepend "?"
			// trim leading "&" if present before prepending it
			if parsedURL.RawQuery != "" {
				methodURL = strings.TrimRight(methodURL, "&") + "&" + renderURLSearchParamsFn
			} else {
				methodURL += "?" + renderURLSearchParamsFn
			}
		}
		return methodURL
	}
}

func buildInitReq(method data.Method) string {
	httpMethod := method.HTTPMethod
	m := `method: "` + httpMethod + `"`
	fields := []string{m}
	if method.HTTPRequestBody == nil || *method.HTTPRequestBody == "*" {
		fields = append(fields, "body: JSON.stringify(req, fm.replacer)")
	} else if *method.HTTPRequestBody != "" {
		fields = append(fields, `body: JSON.stringify(req["`+*method.HTTPRequestBody+`"], fm.replacer)`)
	}

	return strings.Join(fields, ", ")
}

// include is the include template functions copied from copied from:
// https://github.com/helm/helm/blob/8648ccf5d35d682dcd5f7a9c2082f0aaf071e817/pkg/engine/engine.go#L147-L154
func include(t *template.Template) func(name string, data interface{}) (string, error) {
	return func(name string, data interface{}) (string, error) {
		buf := bytes.NewBufferString("")
		if err := t.ExecuteTemplate(buf, name, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
}

func tsTypeKey(r *registry.Registry) func(field *data.Field) string {
	return func(field *data.Field) string {
		// Prefer JsonName if set and different from Name
		name := fieldName(r)(field.Name)
		if r.UseJsonName && field.JsonName != "" && field.JsonName != field.Name {
			name = field.JsonName
		}
		if !r.EmitUnpopulated || field.IsOptional {
			// When EmitUnpopulated is false, the gateway will return undefined for
			// any zero value, so all fields may be undefined. Optional fields, may
			// also be undefined if unset.
			return name + "?"
		}
		// When it is false, only optional fields can be undefined, however they are
		// handled via OneOf style compound types.
		return name
	}
}

func tsTypeDef(r *registry.Registry) func(field *data.Field) string {
	return func(field *data.Field) string {
		t := tsType(r, field)
		info := field.GetType()
		if r.EmitUnpopulated && (!isScalaType(info.Type) || info.IsRepeated) {
			// When EmitUnpopulated is true, zero values will be emitted.
			// Messages and lists may be null.
			return t + " | null"
		}
		return t
	}
}

func tsType(r *registry.Registry, fieldType data.Type) string {
	info := fieldType.GetType()
	typeInfo, ok := r.Types[info.Type]
	if ok && typeInfo.IsMapEntry {
		keyType := tsType(r, typeInfo.KeyType)
		valueType := tsType(r, typeInfo.ValueType)

		return fmt.Sprintf("Record<%s, %s>", keyType, valueType)
	}
	var typeStr string
	switch {
	case mapWellKnownType(info.Type) != "":
		typeStr = mapWellKnownType(info.Type)
	case strings.Index(info.Type, ".") != 0:
		typeStr = mapScalaType(info.Type)
	case !info.IsExternal:
		typeStr = typeInfo.PackageIdentifier
	default:
		typeStr = data.GetModuleName(typeInfo.Package, typeInfo.File) + "." + typeInfo.PackageIdentifier
	}

	if info.IsRepeated {
		typeStr += "[]"
	}
	return typeStr
}

func mapWellKnownType(protoType string) string {
	switch protoType {
	case ".google.protobuf.BoolValue":
		return "boolean | null"
	case ".google.protobuf.StringValue":
		return "string | null"
	case ".google.protobuf.DoubleValue",
		".google.protobuf.FloatValue",
		".google.protobuf.Int32Value",
		".google.protobuf.Int64Value",
		".google.protobuf.UInt32Value",
		".google.protobuf.UInt64Value":
		return "number | null"
	case ".google.protobuf.ListValue":
		return "StructPBValue[]"
	case ".google.protobuf.Struct":
		return "{ [key: string]: StructPBValue }"
	}
	return ""
}

func mapScalaType(protoType string) string {
	switch protoType {
	case "uint64", "sint64", "int64", "fixed64", "sfixed64", "string":
		return "string"
	case "float", "double", "int32", "sint32", "uint32", "fixed32", "sfixed32":
		return "number"
	case "bool":
		return "boolean"
	case "bytes":
		return "Uint8Array"
	}
	return ""
}

func isScalaType(protoType string) bool {
	return mapScalaType(protoType) != ""
}
