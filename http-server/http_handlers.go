package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterHardwareServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				log.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				log.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()
	client := hardware.NewHardwareServiceClient(conn)

	hardwarePushPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "hardware"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("POST", hardwarePushPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		//var hw hardware.Hardware
		var hw util.HardwareWrapper
		newReader, berr := utilities.IOReaderFactory(req.Body)
		if berr != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if err := json.NewDecoder(newReader()).Decode(&hw); err != nil && err != io.EOF {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if _, err := client.Push(ctx, &hardware.PushRequest{Data: hw.Hardware}); err != nil {
			log.Println(err) ///////
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		} else {
			w.Write([]byte("Hardware data pushed successfully\n"))
		}
	})

	hardwareByMACPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"v1", "hardware", "mac"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("POST", hardwareByMACPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var gr hardware.GetRequest
		newReader, berr := utilities.IOReaderFactory(req.Body)
		if berr != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if err := json.NewDecoder(newReader()).Decode(&gr); err != nil && err != io.EOF {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		hw, err := client.ByMAC(context.Background(), &hardware.GetRequest{Mac: gr.Mac})
		if err != nil {
			log.Fatal(err)
		}
		b, err := json.Marshal(util.HardwareWrapper{Hardware: hw})
		if err != nil {
			log.Println(err) ///////
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		w.Write(b)
		w.Write([]byte("\n"))
	})

	hardwareByIPPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"v1", "hardware", "ip"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("POST", hardwareByIPPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var gr hardware.GetRequest
		newReader, berr := utilities.IOReaderFactory(req.Body)
		if berr != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if err := json.NewDecoder(newReader()).Decode(&gr); err != nil && err != io.EOF {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		hw, err := client.ByIP(context.Background(), &hardware.GetRequest{Ip: gr.Ip})
		if err != nil {
			log.Fatal(err)
		}
		b, err := json.Marshal(util.HardwareWrapper{Hardware: hw})
		if err != nil {
			log.Println(err) ///////
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		w.Write(b)
		w.Write([]byte("\n"))
	})

	hardwareByIDPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"v1", "hardware", "id"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", hardwareByIDPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var gr hardware.GetRequest
		val, ok := pathParams["id"]
		if !ok {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "missing parameter %s", "id").Error()))
		}

		gr.Id, err = runtime.String(val)

		if err != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err).Error()))
		}

		hw, err := client.ByID(context.Background(), &hardware.GetRequest{Id: gr.Id})
		if err != nil {
			log.Fatal(err)
		}
		b, err := json.Marshal(util.HardwareWrapper{Hardware: hw})
		if err != nil {
			log.Println(err) ///////
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		w.Write(b)
		w.Write([]byte("\n"))
	})

	hardwareAllPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "hardware"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", hardwareAllPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		alls, err := client.All(context.Background(), &hardware.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		var hw *hardware.Hardware
		err = nil
		for hw, err = alls.Recv(); err == nil && hw != nil; hw, err = alls.Recv() {
			b, err := json.Marshal(util.HardwareWrapper{Hardware: hw})
			if err != nil {
				log.Println(err) ///////
				w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
			}
			w.Write(b)
			w.Write([]byte("\n"))
		}
		if err != nil && err != io.EOF {
			log.Println(err) ///////
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
	})

	return nil
}
