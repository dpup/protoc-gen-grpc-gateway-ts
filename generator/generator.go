package generator

import (
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"text/template"

	"google.golang.org/protobuf/types/pluginpb"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/registry"
	"github.com/pkg/errors"
)

// TypeScriptGRPCGatewayGenerator is the protobuf generator for typescript.
type TypeScriptGRPCGatewayGenerator struct {
	Registry *registry.Registry
}

// New returns an initialised generator.
func New(reg *registry.Registry) (*TypeScriptGRPCGatewayGenerator, error) {
	return &TypeScriptGRPCGatewayGenerator{
		Registry: reg,
	}, nil
}

// Generate take a code generator request and returns a response. it analyse request with registry
// and use the generated data to render ts files.
func (t *TypeScriptGRPCGatewayGenerator) Generate(
	req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	resp := &pluginpb.CodeGeneratorResponse{}

	filesData, err := t.Registry.Analyse(req)
	if err != nil {
		return nil, errors.Wrap(err, "error analysing proto files")
	}
	tmpl := ServiceTemplate(t.Registry)
	slog.Debug("generating files", slog.Any("files", req.GetFileToGenerate()))

	requiresFetchModule := false
	// feed fileData into rendering process
	for _, fileData := range filesData {
		if !t.Registry.IsFileToGenerate(fileData.Name) {
			slog.Debug("file is not the file to generate, skipping", slog.String("fileName", fileData.Name))
			continue
		}

		slog.Debug("generating file", slog.String("fileName", fileData.TSFileName))
		data := &TemplateData{
			File:               fileData,
			EnableStylingCheck: t.Registry.EnableStylingCheck,
			UseStaticClasses:   t.Registry.UseStaticClasses,
		}
		generated, err := t.generateFile(data, tmpl)
		if err != nil {
			return nil, errors.Wrap(err, "error generating file")
		}
		resp.File = append(resp.File, generated)
		requiresFetchModule = requiresFetchModule || fileData.Services.RequiresFetchModule()
	}

	if requiresFetchModule {
		fetchTmpl := FetchModuleTemplate()
		slog.Debug("generate fetch template")
		generatedFetch, err := t.generateFetchModule(fetchTmpl)
		if err != nil {
			return nil, errors.Wrap(err, "error generating fetch module")
		}

		resp.File = append(resp.File, generatedFetch)
	}

	return resp, nil
}

func (t *TypeScriptGRPCGatewayGenerator) generateFile(
	data *TemplateData, tmpl *template.Template) (*pluginpb.CodeGeneratorResponse_File, error) {
	w := bytes.NewBufferString("")

	if data.IsEmpty() {
		//nolint:staticcheck // QF1012: WriteString with Sprintln is readable, optimization not critical
		w.WriteString(fmt.Sprintln("export default {}"))
	} else {
		err := tmpl.Execute(w, data)
		if err != nil {
			return nil, errors.Wrapf(err, "error generating ts file for %s", data.Name)
		}
	}

	fileName := data.TSFileName
	content := strings.TrimSpace(w.String())

	return &pluginpb.CodeGeneratorResponse_File{
		Name:           &fileName,
		InsertionPoint: nil,
		Content:        &content,
	}, nil
}

func (t *TypeScriptGRPCGatewayGenerator) generateFetchModule(
	tmpl *template.Template) (*pluginpb.CodeGeneratorResponse_File, error) {
	w := bytes.NewBufferString("")
	fileName := filepath.Join(t.Registry.FetchModuleDirectory, t.Registry.FetchModuleFilename)
	err := tmpl.Execute(w, &TemplateData{
		EnableStylingCheck: t.Registry.EnableStylingCheck,
		UseStaticClasses:   t.Registry.UseStaticClasses,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error generating fetch module at %s", fileName)
	}

	content := strings.TrimSpace(w.String())
	return &pluginpb.CodeGeneratorResponse_File{
		Name:           &fileName,
		InsertionPoint: nil,
		Content:        &content,
	}, nil
}
