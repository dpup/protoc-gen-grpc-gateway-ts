package generator

import _ "embed"

//go:embed fetchtmpl.ts
var fetchTmplScript string

var fetchTmplHeader = `{{- if not .EnableStylingCheck}}
/* eslint-disable */
// @ts-nocheck
{{- end}}
`

var fetchTmpl = fetchTmplHeader + fetchTmplScript
