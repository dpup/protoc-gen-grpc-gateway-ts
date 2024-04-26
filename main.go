package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/generator"
	"github.com/dpup/protoc-gen-grpc-gateway-ts/registry"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	log "github.com/sirupsen/logrus" // nolint: depguard
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
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

	req := decodeReq()
	paramsMap := getParamsMap(req)
	for k, v := range paramsMap {
		if err := flag.CommandLine.Set(k, v); err != nil {
			return fmt.Errorf("error setting flag %s: %w", k, err)
		}
	}

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
		return fmt.Errorf("error instantiating a new registry: %w", err)
	}

	g, err := generator.New(reg, *enableStylingCheck)
	if err != nil {
		return fmt.Errorf("error instantiating a new generator: %w", err)
	}
	log.Debug("Starts generating file request")
	resp, err := g.Generate(req)
	if err != nil {
		return fmt.Errorf("error generating output: %w", err)
	}
	features := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	resp.SupportedFeatures = &features

	encodeResponse(resp)
	log.Debug("generation finished")

	return nil
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
				return fmt.Errorf("error parsing log level %s: %s", levelStr, err)
			}
			log.SetLevel(level)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	}
	return nil
}

func getParamsMap(req *plugin.CodeGeneratorRequest) map[string]string {
	paramsMap := make(map[string]string)
	params := req.GetParameter()

	for _, p := range strings.Split(params, ",") {
		if i := strings.Index(p, "="); i < 0 {
			paramsMap[p] = ""
		} else {
			paramsMap[p[0:i]] = p[i+1:]
		}
	}

	return paramsMap
}

func decodeReq() *plugin.CodeGeneratorRequest {
	req := &plugin.CodeGeneratorRequest{}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	err = proto.Unmarshal(data, req)
	if err != nil {
		panic(err)
	}
	return req
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
