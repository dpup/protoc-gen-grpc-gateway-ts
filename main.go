package main

import (
	"flag"
	"os"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/generator"
	"github.com/dpup/protoc-gen-grpc-gateway-ts/registry"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus" // nolint: depguard
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	var useProtoNames = flag.Bool("use_proto_names", false, "field names will match the protofile instead of lowerCamelCase")
	var emitUnpopulated = flag.Bool("emit_unpopulated", false, "expect the gRPC Gateway to send zero values over the wire")
	var fetchModuleDirectory = flag.String("fetch_module_directory", ".", "where shared typescript file should be placed, default $(pwd)")
	var fetchModuleFilename = flag.String("fetch_module_filename", "fetch.pb.ts", "name of shard typescript file")
	var tsImportRoots = flag.String("ts_import_roots", "", "defaults to $(pwd)")
	var tsImportRootAliases = flag.String("ts_import_root_aliases", "", "use import aliases instead of relative paths")

	var enableStylingCheck = flag.Bool("enable_styling_check", false, "TODO")

	var logtostderr = flag.Bool("logtostderr", false, "turn on logging to stderr")
	var loglevel = flag.String("loglevel", "info", "defines the logging level. Values are debug, info, warn, error")

	flag.Parse()

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		if err := configureLogging(*logtostderr, *loglevel); err != nil {
			return err
		}

		reg, err := registry.NewRegistry(registry.Options{
			UseProtoNames:        *useProtoNames,
			EmitUnpopulated:      *emitUnpopulated,
			FetchModuleDirectory: *fetchModuleDirectory,
			FetchModuleFileName:  *fetchModuleFilename,
			TSImportRoots:        *tsImportRoots,
			TSImportRootAliases:  *tsImportRootAliases,
		})
		if err != nil {
			return errors.Wrap(err, "error instantiating a new registry")
		}

		g, err := generator.New(reg, *enableStylingCheck)
		if err != nil {
			return errors.Wrap(err, "error instantiating a new generator")
		}

		log.Debug("Starts generating file request")
		resp, err := g.Generate(gen.Request)
		if err != nil {
			return errors.Wrap(err, "error generating output")
		}

		encodeResponse(resp)
		log.Debug("generation finished")

		return nil
	})

}

func encodeResponse(resp proto.Message) {
	data, err := proto.Marshal(resp)
	if err != nil {
		panic(err)
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		panic(err)
	}
}

func configureLogging(enableLogging bool, levelStr string) error {
	if enableLogging {
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp: true,
		})
		log.SetOutput(os.Stderr)
		log.Debugf("Logging configured completed, logging has been enabled")
		if levelStr != "" {
			level, err := log.ParseLevel(levelStr)
			if err != nil {
				return errors.Wrapf(err, "error parsing log level %s", levelStr)
			}
			log.SetLevel(level)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	}
	return nil
}
