package main
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	rpcpb "github.com/iost-official/go-iost/rpc/pb"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strings"
)

var grpcAddr = "0.0.0.0:30002"

func errorHandler(_ context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	w.WriteHeader(400)
	bytes, e := json.Marshal(err)
	if e != nil {
		bytes = []byte(fmt.Sprint(err))
	}
	w.Write(bytes)
}

func startGateway() error {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
		runtime.WithProtoErrorHandler(errorHandler))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := rpcpb.RegisterApiServiceHandlerFromEndpoint(context.Background(), mux,  grpcAddr, opts)
	if err != nil {
		return err
	}
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "HEAD", "POST", "PUT", "DELETE"},
		AllowedOrigins: []string{"*"},
	})
	gatewayServer := &http.Server{
		Addr:    "0.0.0.0:30001",
		Handler: http.HandlerFunc( func(w http.ResponseWriter, req *http.Request) {
			c.Handler(mux).ServeHTTP(w, req)
		}),
	}
	err = gatewayServer.ListenAndServe()
	if err != http.ErrServerClosed {
		panic(err)
	}
	return nil
}

func main() {
	if len(os.Args) >= 2 {
		addr := os.Args[1]
		if !strings.HasSuffix(addr, ":30002") {
			addr += ":30002"
		}
		grpcAddr = addr
	}
	fmt.Println("grpc connect to", grpcAddr)
	var err error
	err = startGateway()
	if err != nil {
		panic(err)
	}
}
