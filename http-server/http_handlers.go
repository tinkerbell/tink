package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	tt "text/template"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"

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

	// hardware push handler | POST /v1/hardware
	hardwarePushPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "hardware"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("POST", hardwarePushPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var hw util.HardwareWrapper
		newReader, berr := utilities.IOReaderFactory(req.Body)
		if berr != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if err := json.NewDecoder(newReader()).Decode(&hw); err != nil && err != io.EOF {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if _, err := client.Push(ctx, &hardware.PushRequest{Data: hw.Hardware}); err != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
			w.Write([]byte("\n"))
			return
		}
		writeResponse(w, "Hardware data pushed successfully")
	})

	// hardware mac handler | POST /v1/hardware/mac
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
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		w.Write(b)
		w.Write([]byte("\n"))
	})

	// hardware ip handler | POST /v1/hardware/ip
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
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		w.Write(b)
		w.Write([]byte("\n"))
	})

	// hardware id handler | GET /v1/hardware/{id}
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
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		w.Write(b)
		w.Write([]byte("\n"))
	})

	// hardware all handler | GET /v1/hardware
	hardwareAllPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "hardware"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", hardwareAllPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		alls, err := client.All(context.Background(), &hardware.Empty{})
		if err != nil {
			writeResponse(w, err.Error())
			return
		}

		var hw *hardware.Hardware
		err = nil
		for hw, err = alls.Recv(); err == nil && hw != nil; hw, err = alls.Recv() {
			b, err := json.Marshal(util.HardwareWrapper{Hardware: hw})
			if err != nil {
				w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
			}
			w.Write(b)
			w.Write([]byte("\n"))
		}
		if err != nil && err != io.EOF {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
	})

	return nil
}

func RegisterTemplateHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
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
	client := template.NewTemplateClient(conn)

	// template create handler | POST /v1/templates
	templateCreatePattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "templates"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("POST", templateCreatePattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var tmpl template.WorkflowTemplate
		newReader, berr := utilities.IOReaderFactory(req.Body)
		if berr != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if err := json.NewDecoder(newReader()).Decode(&tmpl); err != nil && err != io.EOF {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if tmpl.Data != "" {
			if err := tryParseTemplate(tmpl.Data); err != nil {
				w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
				return
			}
			res, err := client.CreateTemplate(context.Background(), &tmpl)
			if err != nil {
				w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
				return
			}
			w.Write([]byte("Created Template: " + res.Id + "\n"))
		}
	})

	// template get handler | GET /v1/templates/{id}
	templateGetPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"v1", "templates", "id"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", templateGetPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var gr template.GetRequest
		val, ok := pathParams["id"]
		if !ok {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "missing parameter %s", "id").Error()))
		}

		gr.Id, err = runtime.String(val)

		t, err := client.GetTemplate(context.Background(), &gr)
		if err != nil {
			writeResponse(w, err.Error())
			return
		}
		writeResponse(w, t.Data)
	})

	// template delete handler | DELETE /v1/templates/{id}
	templateDeletePattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"v1", "templates", "id"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("DELETE", templateDeletePattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var gr template.GetRequest
		val, ok := pathParams["id"]
		if !ok {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "missing parameter %s", "id").Error()))
		}

		gr.Id, err = runtime.String(val)

		if _, err := client.DeleteTemplate(context.Background(), &gr); err != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		w.Write([]byte("Template " + gr.Id + " deleted successfully\n"))
	})

	// template list handler | GET /v1/templates
	templateListPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "templates"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", templateListPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var (
			id        = "Template ID"
			name      = "Template Name"
			createdAt = "Created At"
			updatedAt = "Updated At"
		)

		t := table.NewWriter()
		t.SetOutputMirror(w)
		t.AppendHeader(table.Row{id, name, createdAt, updatedAt})
		list, err := client.ListTemplates(context.Background(), &template.Empty{})
		if err != nil {
			writeResponse(w, err.Error())
			return
		}

		var tmp *template.WorkflowTemplate
		err = nil
		for tmp, err = list.Recv(); err == nil && tmp.Name != ""; tmp, err = list.Recv() {
			cr := *tmp.CreatedAt
			up := *tmp.UpdatedAt
			t.AppendRows([]table.Row{
				{tmp.Id, tmp.Name, time.Unix(cr.Seconds, 0), time.Unix(up.Seconds, 0)},
			})
		}

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		t.Render()
	})

	return nil
}

func RegisterWorkflowSvcHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
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
	client := workflow.NewWorkflowSvcClient(conn)

	// workflow create handler | POST /v1/workflows
	workflowCreatePattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "workflows"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("POST", workflowCreatePattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var gr workflow.GetRequest
		newReader, berr := utilities.IOReaderFactory(req.Body)
		if berr != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		if err := json.NewDecoder(newReader()).Decode(&gr); err != nil && err != io.EOF {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", berr).Error()))
		}

		wf, err := client.GetWorkflow(context.Background(), &gr)
		if err != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error())) ///////
			return
		}
		writeResponse(w, wf.Data)
	})

	// workflow get handler | GET /v1/workflows/{id}
	workflowGetPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"v1", "workflows", "id"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", workflowGetPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var gr workflow.GetRequest
		val, ok := pathParams["id"]
		if !ok {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "missing parameter %s", "id").Error()))
		}

		gr.Id, err = runtime.String(val)

		if err != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err).Error()))
		}

		wf, err := client.GetWorkflow(context.Background(), &gr)
		if err != nil {
			writeResponse(w, err.Error()) /////
			return
		}

		writeResponse(w, wf.Data)
	})

	// workflow delete handler | DELETE /v1/workflows/{id}
	hardwareByIPPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"v1", "workflows", "id"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("DELETE", hardwareByIPPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		gr := workflow.GetRequest{}
		val, ok := pathParams["id"]
		if !ok {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "missing parameter %s", "id").Error()))
		}

		gr.Id, err = runtime.String(val)

		if _, err := client.DeleteWorkflow(context.Background(), &gr); err != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "%v", err).Error()))
		}
		writeResponse(w, fmt.Sprintf("Template %v deleted successfully", gr.Id))

	})

	// workflow list handler | GET /v1/workflows
	workflowListPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"v1", "workflows"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", workflowListPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		if err != nil {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err).Error()))
		}

		list, err := client.ListWorkflows(context.Background(), &workflow.Empty{})
		if err != nil {
			writeResponse(w, err.Error())
		}

		var wf *workflow.Workflow
		err = nil
		for wf, err = list.Recv(); err == nil && wf.Id != ""; wf, err = list.Recv() {
			b, err := json.Marshal(wf)
			if err != nil {
				writeResponse(w, status.Errorf(codes.InvalidArgument, "%v", err).Error())
			}
			writeResponse(w, string(b))
		}

		if err != nil && err != io.EOF {
			writeResponse(w, err.Error())
		}
	})

	// workflow state handler | GET /v1/workflows/{id}/state
	workflowStatePattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2, 2, 3}, []string{"v1", "workflows", "id", "state"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", workflowStatePattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {

		var gr workflow.GetRequest
		val, ok := pathParams["id"]
		if !ok {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "missing parameter %s", "id").Error()))
		}

		gr.Id, err = runtime.String(val)

		wf, err := client.GetWorkflowContext(context.Background(), &gr)
		if err != nil {
			writeResponse(w, err.Error())
		}
		//wfProgress := calWorkflowProgress(wf.CurrentActionIndex, wf.TotalNumberOfActions, wf.CurrentActionState)
		b, err := json.Marshal(wf)
		if err != nil {
			writeResponse(w, status.Errorf(codes.InvalidArgument, "%v", err).Error())
		}
		writeResponse(w, string(b))
	})

	// workflow events handler | GET /v1/workflows/{id}/events
	workflowEventsPattern := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2, 2, 3}, []string{"v1", "workflows", "id", "events"}, "", runtime.AssumeColonVerbOpt(true)))
	mux.Handle("GET", workflowEventsPattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		var gr workflow.GetRequest
		val, ok := pathParams["id"]
		if !ok {
			w.Write([]byte(status.Errorf(codes.InvalidArgument, "missing parameter %s", "id").Error()))
		}

		gr.Id, err = runtime.String(val)

		events, err := client.ShowWorkflowEvents(context.Background(), &gr)
		if err != nil {
			writeResponse(w, err.Error())
			return
		}
		var event *workflow.WorkflowActionStatus
		err = nil
		for event, err = events.Recv(); err == nil && event != nil; event, err = events.Recv() {
			b, err := json.Marshal(event)
			if err != nil {
				writeResponse(w, err.Error())
			}
			writeResponse(w, string(b))
		}
		if err != nil && err != io.EOF {
			writeResponse(w, err.Error())
		}
	})

	return nil
}

func tryParseTemplate(data string) error {
	tmpl := *tt.New("")
	if _, err := tmpl.Parse(data); err != nil {
		return err
	}
	return nil
}

// writeResponse appends a new line after res
func writeResponse(w http.ResponseWriter, res string) {
	if _, err := w.Write([]byte(fmt.Sprintln(res))); err != nil {
		logger.Error(err)
	}
}
