package httpserver

import (
	"context"
	"encoding/json"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"github.com/tinkerbell/tink/protos/hardware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net/http"
)

func RegisterHardwareServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	client := hardware.NewHardwareServiceClient(conn)
	hardwarePushPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "hardware"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("POST", hardwarePushPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var hw hardware.Hardware
		newReader, berr := utilities.IOReaderFactory(req.Body)
		if berr != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if err := json.NewDecoder(newReader()).Decode(&hw); err != nil && err != io.EOF {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if _, err := client.Push(ctx, &hardware.PushRequest{Data: &hw}); err != nil {
			log.Println(err)
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		w.Write([]byte("Hardware data pushed successfully\n"))
	})
	return nil
}
