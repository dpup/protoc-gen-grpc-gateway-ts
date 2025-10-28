package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/dpup/protoc-gen-grpc-gateway-ts/generator"
	"github.com/dpup/protoc-gen-grpc-gateway-ts/registry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var (
		useProtoNames    = flag.Bool("use_proto_names", false, "field names will match the proto file")
		useJsonName      = flag.Bool("use_json_name", false, "field names will match the json name")
		useStaticClasses = flag.Bool("use_static_classes", true, "use static classes rather than functions and a client")
		emitUnpopulated  = flag.Bool("emit_unpopulated", false, "expect the gRPC Gateway to send zero values over the wire")

		fetchModuleDirectory = flag.String("fetch_module_directory", ".", "where shared typescript file should be placed")
		fetchModuleFilename  = flag.String("fetch_module_filename", "fetch.pb.ts", "name of shard typescript file")
		tsImportRoots        = flag.String("ts_import_roots", "", "defaults to $(pwd)")
		tsImportRootAliases  = flag.String("ts_import_root_aliases", "", "use import aliases instead of relative paths")

		enableStylingCheck = flag.Bool("enable_styling_check", false, "TODO")

		logtostderr = flag.Bool("logtostderr", false, "turn on logging to stderr")
		loglevel    = flag.String("loglevel", "info", "defines the logging level. Values are debug, info, warn, error")
	)

	flag.Parse()

	req := decodeReq()

	paramsMap := getParamsMap(req)
	for k, v := range paramsMap {
		if k != "" {
			if err := flag.CommandLine.Set(k, v); err != nil {
				return fmt.Errorf("error setting flag %s: %w  [%+v]", k, err, paramsMap)
			}
		}
	}

	if err := configureLogging(*logtostderr, *loglevel); err != nil {
		return err
	}

	reg, err := registry.NewRegistry(registry.Options{
		UseProtoNames:        *useProtoNames,
		UseJsonName:          *useJsonName,
		UseStaticClasses:     *useStaticClasses,
		EnableStylingCheck:   *enableStylingCheck,
		EmitUnpopulated:      *emitUnpopulated,
		FetchModuleDirectory: *fetchModuleDirectory,
		FetchModuleFilename:  *fetchModuleFilename,
		TSImportRoots:        *tsImportRoots,
		TSImportRootAliases:  *tsImportRootAliases,
	})
	if err != nil {
		return fmt.Errorf("error instantiating a new registry: %w", err)
	}

	g, err := generator.New(reg)
	if err != nil {
		return fmt.Errorf("error instantiating a new generator: %w", err)
	}

	slog.Debug("Starts generating file request")

	resp, err := g.Generate(req)
	if err != nil {
		return fmt.Errorf("error generating output: %w", err)
	}

	features := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	resp.SupportedFeatures = &features

	encodeResponse(resp)

	slog.Debug("generation finished")

	return nil
}

func configureLogging(enableLogging bool, levelStr string) error {
	if enableLogging {
		level := slog.LevelInfo
		if levelStr != "" {
			switch levelStr {
			case "debug":
				level = slog.LevelDebug
			case "info":
				level = slog.LevelInfo
			case "warn":
				level = slog.LevelWarn
			case "error":
				level = slog.LevelError
			default:
				return fmt.Errorf("invalid log level %s", levelStr)
			}
		}
		opts := &slog.HandlerOptions{Level: level}
		logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
		slog.SetDefault(logger)
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}
	return nil
}

func getParamsMap(req *pluginpb.CodeGeneratorRequest) map[string]string {
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

func decodeReq() *pluginpb.CodeGeneratorRequest {
	req := &pluginpb.CodeGeneratorRequest{}
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
