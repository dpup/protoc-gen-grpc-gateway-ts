package registry

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/data"
)

func getHTTPAnnotation(m *descriptorpb.MethodDescriptorProto) *annotations.HttpRule {
	option := proto.GetExtension(m.GetOptions(), annotations.E_Http)
	return option.(*annotations.HttpRule)
}

func hasHTTPAnnotation(m *descriptorpb.MethodDescriptorProto) bool {
	return getHTTPAnnotation(m) != nil
}

func extractHTTPMethodPath(rule *annotations.HttpRule) (string, string) {
	pattern := rule.Pattern
	switch pattern.(type) {
	case *annotations.HttpRule_Get:
		return "GET", rule.GetGet()
	case *annotations.HttpRule_Post:
		return "POST", rule.GetPost()
	case *annotations.HttpRule_Put:
		return "PUT", rule.GetPut()
	case *annotations.HttpRule_Patch:
		return "PATCH", rule.GetPatch()
	case *annotations.HttpRule_Delete:
		return "DELETE", rule.GetDelete()
	default:
		panic(fmt.Sprintf("unsupported HTTP method %T", pattern))
	}
}

func getHTTPBody(m *descriptorpb.MethodDescriptorProto) *string {
	if !hasHTTPAnnotation(m) {
		return nil
	}
	rule := getHTTPAnnotation(m)
	return extractHTTPBody(rule)
}

func extractHTTPBody(rule *annotations.HttpRule) *string {
	empty := ""
	pattern := rule.Pattern
	switch pattern.(type) {
	case *annotations.HttpRule_Get:
		return &empty
	default:
		body := rule.GetBody()
		return &body
	}
}

// generateTSMethodName generates a unique TypeScript method name for a binding
func generateTSMethodName(rpcName, httpMethod string, bindingIndex int, existingMethods []*data.Method) string {
	// Primary binding uses the RPC method name as-is
	if bindingIndex == 0 {
		return rpcName
	}

	// For additional bindings, create a name by appending the HTTP method
	// Capitalize first letter of HTTP method (e.g., "post" -> "Post")
	titleCaser := cases.Title(language.English)
	httpMethodTitle := titleCaser.String(strings.ToLower(httpMethod))
	baseName := rpcName + httpMethodTitle

	// Check for conflicts with existing method names
	conflictCount := 0
	for _, method := range existingMethods {
		if method.TSMethodName == baseName || strings.HasPrefix(method.TSMethodName, baseName) {
			conflictCount++
		}
	}

	// If there's a conflict, append a number
	if conflictCount > 0 {
		return fmt.Sprintf("%s%d", baseName, conflictCount+1)
	}

	return baseName
}

// createMethodFromRule creates a Method data structure from an HttpRule
func createMethodFromRule(
	method *descriptorpb.MethodDescriptorProto,
	rule *annotations.HttpRule,
	bindingIndex int,
	serviceData *data.Service,
	inputTypeFQName, outputTypeFQName string,
	isInputTypeExternal, isOutputTypeExternal bool,
) *data.Method {
	httpMethod, url := extractHTTPMethodPath(rule)
	if httpMethod == "" || url == "" {
		// Should not happen for valid rules, but fallback to defaults
		httpMethod = "POST"
		url = "/" + method.GetName()
	}
	body := extractHTTPBody(rule)

	// Generate TypeScript method name
	tsMethodName := generateTSMethodName(method.GetName(), httpMethod, bindingIndex, serviceData.Methods)

	return &data.Method{
		Name: method.GetName(),
		URL:  url,
		Input: &data.MethodArgument{
			Type:       inputTypeFQName,
			IsExternal: isInputTypeExternal,
		},
		Output: &data.MethodArgument{
			Type:       outputTypeFQName,
			IsExternal: isOutputTypeExternal,
		},
		ServerStreaming: method.GetServerStreaming(),
		ClientStreaming: method.GetClientStreaming(),
		HTTPMethod:      httpMethod,
		HTTPRequestBody: body,
		BindingIndex:    bindingIndex,
		TSMethodName:    tsMethodName,
	}
}

func (r *Registry) analyseService(
	fileData *data.File,
	packageName, fileName string,
	service *descriptorpb.ServiceDescriptorProto) {
	packageIdentifier := service.GetName()
	fqName := "." + packageName + "." + packageIdentifier

	// register itself in the registry map
	r.Types[fqName] = &TypeInformation{
		FullyQualifiedName: fqName,
		Package:            packageName,
		File:               fileName,
		PackageIdentifier:  packageIdentifier,
		LocalIdentifier:    service.GetName(),
	}

	serviceData := data.NewService()
	serviceData.Name = service.GetName()
	serviceURLPart := packageName + "." + serviceData.Name

	for _, method := range service.Method {
		// don't support client streaming, will ignore the client streaming method
		if method.GetClientStreaming() {
			continue
		}

		inputTypeFQName := *method.InputType
		isInputTypeExternal := r.isExternalDependenciesOutsidePackage(inputTypeFQName, packageName)

		if isInputTypeExternal {
			fileData.ExternalDependingTypes = append(fileData.ExternalDependingTypes, inputTypeFQName)
		}

		outputTypeFQName := *method.OutputType
		isOutputTypeExternal := r.isExternalDependenciesOutsidePackage(outputTypeFQName, packageName)

		if isOutputTypeExternal {
			fileData.ExternalDependingTypes = append(fileData.ExternalDependingTypes, outputTypeFQName)
		}

		// Process primary HTTP binding
		if hasHTTPAnnotation(method) {
			rule := getHTTPAnnotation(method)

			// Create method for primary binding
			methodData := createMethodFromRule(
				method,
				rule,
				0, // bindingIndex = 0 for primary
				serviceData,
				inputTypeFQName,
				outputTypeFQName,
				isInputTypeExternal,
				isOutputTypeExternal,
			)

			fileData.TrackPackageNonScalarType(methodData.Input)
			fileData.TrackPackageNonScalarType(methodData.Output)

			serviceData.Methods = append(serviceData.Methods, methodData)

			// Process additional bindings
			for idx, additionalRule := range rule.GetAdditionalBindings() {
				additionalMethodData := createMethodFromRule(
					method,
					additionalRule,
					idx+1, // bindingIndex starts at 1 for additional bindings
					serviceData,
					inputTypeFQName,
					outputTypeFQName,
					isInputTypeExternal,
					isOutputTypeExternal,
				)

				fileData.TrackPackageNonScalarType(additionalMethodData.Input)
				fileData.TrackPackageNonScalarType(additionalMethodData.Output)

				serviceData.Methods = append(serviceData.Methods, additionalMethodData)
			}
		} else {
			// No HTTP annotation - use defaults (backward compatibility)
			httpMethod := "POST"
			url := "/" + serviceURLPart + "/" + method.GetName()
			body := getHTTPBody(method)

			methodData := &data.Method{
				Name: method.GetName(),
				URL:  url,
				Input: &data.MethodArgument{
					Type:       inputTypeFQName,
					IsExternal: isInputTypeExternal,
				},
				Output: &data.MethodArgument{
					Type:       outputTypeFQName,
					IsExternal: isOutputTypeExternal,
				},
				ServerStreaming: method.GetServerStreaming(),
				ClientStreaming: method.GetClientStreaming(),
				HTTPMethod:      httpMethod,
				HTTPRequestBody: body,
				BindingIndex:    0,
				TSMethodName:    method.GetName(),
			}

			fileData.TrackPackageNonScalarType(methodData.Input)
			fileData.TrackPackageNonScalarType(methodData.Output)

			serviceData.Methods = append(serviceData.Methods, methodData)
		}
	}

	fileData.Services = append(fileData.Services, serviceData)
}
