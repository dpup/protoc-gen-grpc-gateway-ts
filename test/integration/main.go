package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

// preflightHandler adds the necessary headers in order to serve
// CORS from any origin using the methods "GET", "HEAD", "POST", "PUT", "DELETE"
// We insist, don't do this without consideration in production systems.
//nolint:unparam // r is required by http.HandlerFunc signature
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
}

// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

const endpoint = "localhost:9000"

func main() {
	useProtoNames := flag.Bool("use_proto_names", false, "tell server to use the original proto name in jsonpb")
	emitUnpopulated := flag.Bool("emit_unpopulated", false, "tell server to emit zero values")

	flag.Parse()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcListener, err := net.Listen("tcp4", endpoint)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting server with use_proto_names=%v and emit_unpopulated=%v\n", *useProtoNames, *emitUnpopulated)

	grpcServer := grpc.NewServer()
	RegisterCounterServiceServer(grpcServer, &RealCounterService{})

	gateway := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
		Marshaler: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   *useProtoNames,
				EmitUnpopulated: *emitUnpopulated,
			},
		},
	}))

	err = RegisterCounterServiceHandlerFromEndpoint(ctx, gateway, endpoint, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	})
	if err != nil {
		panic(err)
	}

	go func() {
		defer grpcServer.GracefulStop()
		<-ctx.Done()
	}()

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			panic(err)
		}
	}()

	//nolint:gosec // G114: Test server doesn't need production timeouts
	if err = http.ListenAndServe("localhost:8081", allowCORS(gateway)); err != nil {
		panic(err)
	}

}
