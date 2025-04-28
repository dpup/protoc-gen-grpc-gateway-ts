package registry

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/data"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// importRootSeparator separates the ts import root inside ts_import_roots & ts_import_root_aliases.
const importRootSeparator = ";"

type Options struct {
	// TSImportRootParamsKey contains the key for common_import_root in parameters
	TSImportRoots string
	// TSImportRootAliasParamsKey contains the key for common_import_root_alias in parameters
	TSImportRootAliases string
	// FetchModuleDirectory is the parameter for directory where fetch module will live
	FetchModuleDirectory string
	// FetchModuleFilename is the file name for the individual fetch module
	FetchModuleFilename string
	// UseProtoNames will generate field names the same as defined in the proto
	UseProtoNames bool
	// UseJsonName will generate field names the same as defined in the json_name annotation
	UseJsonName bool
	// UseStaticClasses will cause the generator to generate a static class in the form ServiceName.MethodName, which is
	// the legacy behavior for this generator. If set to false, the generator will generate a client class with methods
	// as well as static methods exported for each service method.
	UseStaticClasses bool
	// EmitUnpopulated mirrors the grpc gateway protojson configuration of the same name and allows
	// clients to differentiate between zero values and optional values that aren't set.
	EmitUnpopulated bool
	// EnableStylingCheck enables both eslint and tsc check for the generated code
	EnableStylingCheck bool
}

// Registry analyze generation request, spits out the data the the rendering process
// it also holds the information about all the types.
type Registry struct {
	Options

	// Types stores the type information keyed by the fully qualified name of a type
	Types map[string]*TypeInformation

	// FilesToGenerate contains a list of actual file to generate, different from all the files from
	// the request, some of which are import files
	FilesToGenerate map[string]bool

	// TSImportRoots represents the ts import root for the generator to figure out required import
	// path, will default to cwd
	TSImportRoots []string

	// TSImportRootAliases if not empty will substitutes the common import root when writing the
	// import into the js file
	TSImportRootAliases []string

	// TSPackages stores the package name keyed by the TS file name
	TSPackages map[string]string
}

// NewRegistry initialise the registry and return the instance.
func NewRegistry(opts Options) (*Registry, error) {
	tsImportRoots, tsImportRootAliases, err := getTSImportRootInformation(opts)
	slog.Debug("found ts import roots", slog.Any("importRoots", tsImportRoots))
	slog.Debug("found ts import root aliases", slog.Any("importRootAliases", tsImportRootAliases))
	if err != nil {
		return nil, errors.Wrap(err, "error getting common import root information")
	}

	slog.Debug("found fetch module directory", slog.String("moduleDir", opts.FetchModuleDirectory))
	slog.Debug("found fetch module name", slog.String("moduleName", opts.FetchModuleFilename))

	return &Registry{
		Options:             opts,
		Types:               make(map[string]*TypeInformation),
		TSPackages:          make(map[string]string),
		TSImportRoots:       tsImportRoots,
		TSImportRootAliases: tsImportRootAliases,
	}, nil
}

func getTSImportRootInformation(opts Options) ([]string, []string, error) {
	tsImportRootsValue := opts.TSImportRoots
	if tsImportRootsValue == "" {
		tsImportRootsValue = "."
	}

	splittedImportRoots := strings.Split(tsImportRootsValue, importRootSeparator)
	numImportRoots := len(splittedImportRoots)

	tsImportRoots := make([]string, 0, numImportRoots)

	for _, r := range splittedImportRoots {
		tsImportRoot := r
		if !path.IsAbs(tsImportRoot) {
			absPath, err := filepath.Abs(tsImportRoot)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error turning path %s into absolute path", tsImportRoot)
			}

			tsImportRoot = absPath
		}

		tsImportRoots = append(tsImportRoots, tsImportRoot)
	}

	tsImportRootAliasValue := opts.TSImportRootAliases
	splittedImportRootAliases := strings.Split(tsImportRootAliasValue, importRootSeparator)
	tsImportRootAliases := make([]string, numImportRoots)
	for i, ra := range splittedImportRootAliases {
		if i >= numImportRoots {
			// in case we have more root alias than root, we will just take the number matches the roots
			break
		}
		tsImportRootAliases[i] = ra
	}

	return tsImportRoots, tsImportRootAliases, nil
}

// TypeInformation store the information about a given type.
type TypeInformation struct {
	// Fully qualified name of the type, it starts with `.` and followed by packages and the nested
	// structure path.
	FullyQualifiedName string
	// Package is the package of the type it belongs to
	Package string
	// Files is the file of the type belongs to, this is important in Typescript as modules is the
	// namespace for types defined inside
	File string
	// ModuleIdentifier is the identifier of the type inside the package, this will be useful for enum
	// and nested enum.
	PackageIdentifier string
	// LocalIdentifier is the identifier inside the types local scope
	LocalIdentifier string
	// ProtoType is the type inside the proto. This is used to tell whether it's an enum or a message
	ProtoType descriptorpb.FieldDescriptorProto_Type
	// IsMapEntry indicates whether this type is a Map Entry
	IsMapEntry bool
	// KeyType is the type information for the map key
	KeyType *data.MapEntryType
	// Value type is the type information for the map value
	ValueType *data.MapEntryType
}

// IsFileToGenerate contains the file to be generated in the request.
func (r *Registry) IsFileToGenerate(name string) bool {
	result, ok := r.FilesToGenerate[name]
	return ok && result
}

// Analyse analyses the the file inputs, stores types information and spits out the rendering data.
func (r *Registry) Analyse(req *pluginpb.CodeGeneratorRequest) (map[string]*data.File, error) {
	r.FilesToGenerate = make(map[string]bool)
	for _, f := range req.GetFileToGenerate() {
		r.FilesToGenerate[f] = true
	}

	files := req.GetProtoFile()
	slog.Debug("about to start anaylyse files", slog.Int("count", len(files)))
	data := make(map[string]*data.File)
	// analyse all files in the request first
	for _, f := range files {
		fileData, err := r.analyseFile(f)
		if err != nil {
			return nil, errors.Wrapf(err, "error analysing file %s", *f.Name)
		}
		data[f.GetName()] = fileData
	}

	// when finishes we have a full map of types and where they are located
	// collect all the external dependencies and back fill it to the file data.
	err := r.collectExternalDependenciesFromData(data)
	if err != nil {
		return nil, errors.Wrap(err, "error collecting external dependency information after analysis finished")
	}

	return data, nil
}

// This simply just concats the parents name and the entity name.
func (r *Registry) getNameOfPackageLevelIdentifier(parents []string, name string) string {
	return strings.Join(parents, "") + name
}

func (r *Registry) getFullQualifiedName(packageName string, parents []string, name string) string {
	additionalNames := 2
	namesToConcat := make([]string, 0, additionalNames+len(parents))

	if packageName != "" {
		namesToConcat = append(namesToConcat, packageName)
	}

	if len(parents) > 0 {
		namesToConcat = append(namesToConcat, parents...)
	}

	namesToConcat = append(namesToConcat, name)

	return "." + strings.Join(namesToConcat, ".")
}

func (r *Registry) isExternalDependenciesOutsidePackage(fqTypeName, packageName string) bool {
	return strings.Index(fqTypeName, "."+packageName) != 0 && strings.Index(fqTypeName, ".") == 0
}

// findRootAliasForPath iterate through all ts_import_roots and try to find an alias with the first
// matching the ts_import_root.
func (r *Registry) findRootAliasForPath(
	predicate func(root string) (bool, error)) (string, string, error) {
	foundAtRoot := ""
	alias := ""
	for i, root := range r.TSImportRoots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return "", "", errors.Wrapf(err, "error looking up absolute path for %s", err)
		}

		found, err := predicate(absRoot)
		if err != nil {
			return "", "", errors.Wrapf(err, "error verifying the root %s for", absRoot)
		}

		if found {
			foundAtRoot = root
			if i >= len(r.TSImportRootAliases) {
				alias = ""
			} else {
				alias = r.TSImportRootAliases[i]
			}

			break
		}
	}

	return foundAtRoot, alias, nil
}

// getSourceFileForImport will return source file for import use.
// if alias is provided it will try to replace the absolute root with target's absolute path with alias
// if no alias then it will try to return a relative path to the source file.
func (r *Registry) getSourceFileForImport(source, target, root, alias string) (string, error) {
	var ret string
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", errors.Wrapf(err, "error looking up absolute path for target %s", target)
	}

	// if an alias has been provided, that means there's no need to get relative path
	if alias != "" {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return "", errors.Wrapf(err, "error looking up absolute path for root %s", root)
		}

		ret = strings.ReplaceAll(absTarget, absRoot, alias)
		slog.Debug(fmt.Sprintf("replacing root alias %s for %s, result: %s", alias, target, ret))
	} else { // return relative path here
		slog.Debug("no root alias found, trying to get the relative path", slog.String("target", target))
		absSource, err := filepath.Abs(source)
		if err != nil {
			return "", errors.Wrapf(err, "error looking up absolute directory with base dir: %s", source)
		}

		ret, err = filepath.Rel(filepath.Dir(absSource), absTarget)
		if err != nil {
			return "", errors.Wrapf(err, "error looking up relative path for source target %s", target)
		}

		slashPath := filepath.ToSlash(ret)
		slog.Debug(fmt.Sprintf("got relative path %s for %s", target, slashPath))

		// sub directory will not have relative path ./, if this happens, prepend one
		if !strings.HasPrefix(slashPath, "../") {
			ret = filepath.FromSlash("./" + slashPath)
		}

		slog.Debug(fmt.Sprintf("no root alias found, trying to get the relative path for %s, result: %s", target, ret))
	}

	// remove .ts suffix if there's any
	suffixIndex := strings.LastIndex(ret, ".ts")
	if suffixIndex != -1 {
		ret = ret[0:suffixIndex]
	}

	return ret, nil
}

func (r *Registry) collectExternalDependenciesFromData(filesData map[string]*data.File) error {
	for _, fileData := range filesData {
		slog.Debug("collecting dependencies information", slog.String("fileName", fileData.TSFileName))
		// dependency group up the dependency by package+file
		dependencies := make(map[string]*data.Dependency)
		for _, typeName := range fileData.ExternalDependingTypes {
			typeInfo, ok := r.Types[typeName]
			if !ok {
				return errors.Errorf("cannot find type info for %s, $v", typeName)
			}
			if typeInfo.File == "google/protobuf/wrappers.proto" {
				// Skip well-known wrapper types without importing them as an external dependency,
				// since their types are converted to native TypeScript types by mapWellKnownType.
				continue
			}
			identifier := typeInfo.Package + "|" + typeInfo.File

			if _, ok := dependencies[identifier]; !ok {
				// only fill in if this file has not been mentioned before.
				// the way import in the generated file works is like
				// import * as [ModuleIdentifier] from '[Source File]'
				// so there only needs to be added once.
				// Referencing types will be [ModuleIdentifier].[PackageIdentifier]
				base := fileData.TSFileName
				target := data.GetTSFileName(typeInfo.File)
				var sourceFile string
				if pkg, ok := r.TSPackages[target]; ok {
					slog.Debug("package import override has been found", slog.String("pkg", pkg), slog.String("target", target))
					sourceFile = pkg
				} else {
					foundAtRoot, alias, err := r.findRootAliasForPath(func(absRoot string) (bool, error) {
						completePath := filepath.Join(absRoot, typeInfo.File)
						_, err := os.Stat(completePath)
						if err != nil {
							if os.IsNotExist(err) {
								return false, nil
							}
							return false, err
						}
						return true, nil
					})
					if err != nil {
						return errors.WithStack(err)
					}

					if foundAtRoot != "" {
						target = filepath.Join(foundAtRoot, target)
					}

					sourceFile, err = r.getSourceFileForImport(base, target, foundAtRoot, alias)
					if err != nil {
						return errors.Wrap(err, "error getting source file for import")
					}
				}
				dependencies[identifier] = &data.Dependency{
					ModuleIdentifier: data.GetModuleName(typeInfo.Package, typeInfo.File),
					SourceFile:       sourceFile,
				}
			}
		}

		for _, dependency := range dependencies {
			fileData.AddDependency(dependency)
		}
	}

	return nil
}
