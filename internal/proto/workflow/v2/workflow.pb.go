// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: internal/proto/workflow/v2/workflow.proto

package workflow

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StreamWorkflowsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AgentId string `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"`
}

func (x *StreamWorkflowsRequest) Reset() {
	*x = StreamWorkflowsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamWorkflowsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamWorkflowsRequest) ProtoMessage() {}

func (x *StreamWorkflowsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamWorkflowsRequest.ProtoReflect.Descriptor instead.
func (*StreamWorkflowsRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{0}
}

func (x *StreamWorkflowsRequest) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

type StreamWorkflowsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Workflow *Workflow `protobuf:"bytes,1,opt,name=workflow,proto3" json:"workflow,omitempty"`
}

func (x *StreamWorkflowsResponse) Reset() {
	*x = StreamWorkflowsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamWorkflowsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamWorkflowsResponse) ProtoMessage() {}

func (x *StreamWorkflowsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamWorkflowsResponse.ProtoReflect.Descriptor instead.
func (*StreamWorkflowsResponse) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{1}
}

func (x *StreamWorkflowsResponse) GetWorkflow() *Workflow {
	if x != nil {
		return x.Workflow
	}
	return nil
}

type PublishEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Event *Event `protobuf:"bytes,1,opt,name=event,proto3" json:"event,omitempty"`
}

func (x *PublishEventRequest) Reset() {
	*x = PublishEventRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublishEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishEventRequest) ProtoMessage() {}

func (x *PublishEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishEventRequest.ProtoReflect.Descriptor instead.
func (*PublishEventRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{2}
}

func (x *PublishEventRequest) GetEvent() *Event {
	if x != nil {
		return x.Event
	}
	return nil
}

type PublishEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PublishEventResponse) Reset() {
	*x = PublishEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublishEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishEventResponse) ProtoMessage() {}

func (x *PublishEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishEventResponse.ProtoReflect.Descriptor instead.
func (*PublishEventResponse) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{3}
}

type Workflow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique identifier for a workflow.
	WorkflowId string `protobuf:"bytes,1,opt,name=workflow_id,json=workflowId,proto3" json:"workflow_id,omitempty"`
	// The actions that make up the workflow.
	Actions []*Workflow_Action `protobuf:"bytes,2,rep,name=actions,proto3" json:"actions,omitempty"`
}

func (x *Workflow) Reset() {
	*x = Workflow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Workflow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Workflow) ProtoMessage() {}

func (x *Workflow) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Workflow.ProtoReflect.Descriptor instead.
func (*Workflow) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{4}
}

func (x *Workflow) GetWorkflowId() string {
	if x != nil {
		return x.WorkflowId
	}
	return ""
}

func (x *Workflow) GetActions() []*Workflow_Action {
	if x != nil {
		return x.Actions
	}
	return nil
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique identifier for a workflow.
	WorkflowId string `protobuf:"bytes,1,opt,name=workflow_id,json=workflowId,proto3" json:"workflow_id,omitempty"`
	// Additional data that compliments the event type.
	//
	// Types that are assignable to Event:
	//
	//	*Event_ActionStarted_
	//	*Event_ActionSucceeded_
	//	*Event_ActionFailed_
	Event isEvent_Event `protobuf_oneof:"event"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{5}
}

func (x *Event) GetWorkflowId() string {
	if x != nil {
		return x.WorkflowId
	}
	return ""
}

func (m *Event) GetEvent() isEvent_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (x *Event) GetActionStarted() *Event_ActionStarted {
	if x, ok := x.GetEvent().(*Event_ActionStarted_); ok {
		return x.ActionStarted
	}
	return nil
}

func (x *Event) GetActionSucceeded() *Event_ActionSucceeded {
	if x, ok := x.GetEvent().(*Event_ActionSucceeded_); ok {
		return x.ActionSucceeded
	}
	return nil
}

func (x *Event) GetActionFailed() *Event_ActionFailed {
	if x, ok := x.GetEvent().(*Event_ActionFailed_); ok {
		return x.ActionFailed
	}
	return nil
}

type isEvent_Event interface {
	isEvent_Event()
}

type Event_ActionStarted_ struct {
	ActionStarted *Event_ActionStarted `protobuf:"bytes,2,opt,name=action_started,json=actionStarted,proto3,oneof"`
}

type Event_ActionSucceeded_ struct {
	ActionSucceeded *Event_ActionSucceeded `protobuf:"bytes,3,opt,name=action_succeeded,json=actionSucceeded,proto3,oneof"`
}

type Event_ActionFailed_ struct {
	ActionFailed *Event_ActionFailed `protobuf:"bytes,4,opt,name=action_failed,json=actionFailed,proto3,oneof"`
}

func (*Event_ActionStarted_) isEvent_Event() {}

func (*Event_ActionSucceeded_) isEvent_Event() {}

func (*Event_ActionFailed_) isEvent_Event() {}

type PublishResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PublishResponse) Reset() {
	*x = PublishResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublishResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishResponse) ProtoMessage() {}

func (x *PublishResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishResponse.ProtoReflect.Descriptor instead.
func (*PublishResponse) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{6}
}

type Workflow_Action struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique identifier for an action in the context of a workflow.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The name of the action. This can be used to identify actions in logging.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// The image to run.
	Image string `protobuf:"bytes,3,opt,name=image,proto3" json:"image,omitempty"`
	// The command to execute when launching the image. When using Docker as the action runtime
	// it is used as the entrypoint.
	Cmd *string `protobuf:"bytes,4,opt,name=cmd,proto3,oneof" json:"cmd,omitempty"`
	// Arguments to pass to the container.
	Args []string `protobuf:"bytes,5,rep,name=args,proto3" json:"args,omitempty"`
	// Environment variables to configure when launching the container.
	Env map[string]string `protobuf:"bytes,6,rep,name=env,proto3" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Volumes to mount when launching the container.
	Volumes []string `protobuf:"bytes,7,rep,name=volumes,proto3" json:"volumes,omitempty"`
	// The network namespace to launch the container in.
	NetworkNamespace *string `protobuf:"bytes,8,opt,name=network_namespace,json=networkNamespace,proto3,oneof" json:"network_namespace,omitempty"`
}

func (x *Workflow_Action) Reset() {
	*x = Workflow_Action{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Workflow_Action) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Workflow_Action) ProtoMessage() {}

func (x *Workflow_Action) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Workflow_Action.ProtoReflect.Descriptor instead.
func (*Workflow_Action) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{4, 0}
}

func (x *Workflow_Action) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Workflow_Action) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Workflow_Action) GetImage() string {
	if x != nil {
		return x.Image
	}
	return ""
}

func (x *Workflow_Action) GetCmd() string {
	if x != nil && x.Cmd != nil {
		return *x.Cmd
	}
	return ""
}

func (x *Workflow_Action) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

func (x *Workflow_Action) GetEnv() map[string]string {
	if x != nil {
		return x.Env
	}
	return nil
}

func (x *Workflow_Action) GetVolumes() []string {
	if x != nil {
		return x.Volumes
	}
	return nil
}

func (x *Workflow_Action) GetNetworkNamespace() string {
	if x != nil && x.NetworkNamespace != nil {
		return *x.NetworkNamespace
	}
	return ""
}

type Event_ActionStarted struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique identifier for an action in the context of a workflow.
	ActionId string `protobuf:"bytes,1,opt,name=action_id,json=actionId,proto3" json:"action_id,omitempty"`
}

func (x *Event_ActionStarted) Reset() {
	*x = Event_ActionStarted{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event_ActionStarted) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event_ActionStarted) ProtoMessage() {}

func (x *Event_ActionStarted) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event_ActionStarted.ProtoReflect.Descriptor instead.
func (*Event_ActionStarted) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{5, 0}
}

func (x *Event_ActionStarted) GetActionId() string {
	if x != nil {
		return x.ActionId
	}
	return ""
}

type Event_ActionSucceeded struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique identifier for an action in the context of a workflow.
	ActionId string `protobuf:"bytes,1,opt,name=action_id,json=actionId,proto3" json:"action_id,omitempty"`
}

func (x *Event_ActionSucceeded) Reset() {
	*x = Event_ActionSucceeded{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event_ActionSucceeded) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event_ActionSucceeded) ProtoMessage() {}

func (x *Event_ActionSucceeded) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event_ActionSucceeded.ProtoReflect.Descriptor instead.
func (*Event_ActionSucceeded) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{5, 1}
}

func (x *Event_ActionSucceeded) GetActionId() string {
	if x != nil {
		return x.ActionId
	}
	return ""
}

type Event_ActionFailed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique identifier for an action in the context of a workflow.
	ActionId string `protobuf:"bytes,1,opt,name=action_id,json=actionId,proto3" json:"action_id,omitempty"`
	// A UpperCamelCase word or phrase concisly describing why an action failed. It is typically
	// provided by the action itself.
	FailureReason *string `protobuf:"bytes,2,opt,name=failure_reason,json=failureReason,proto3,oneof" json:"failure_reason,omitempty"`
	// A free-form human readable string elaborating on the reason for failure. It is typically
	// provided by the action itself.
	FailureMessage *string `protobuf:"bytes,3,opt,name=failure_message,json=failureMessage,proto3,oneof" json:"failure_message,omitempty"`
}

func (x *Event_ActionFailed) Reset() {
	*x = Event_ActionFailed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event_ActionFailed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event_ActionFailed) ProtoMessage() {}

func (x *Event_ActionFailed) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_workflow_v2_workflow_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event_ActionFailed.ProtoReflect.Descriptor instead.
func (*Event_ActionFailed) Descriptor() ([]byte, []int) {
	return file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP(), []int{5, 2}
}

func (x *Event_ActionFailed) GetActionId() string {
	if x != nil {
		return x.ActionId
	}
	return ""
}

func (x *Event_ActionFailed) GetFailureReason() string {
	if x != nil && x.FailureReason != nil {
		return *x.FailureReason
	}
	return ""
}

func (x *Event_ActionFailed) GetFailureMessage() string {
	if x != nil && x.FailureMessage != nil {
		return *x.FailureMessage
	}
	return ""
}

var File_internal_proto_workflow_v2_workflow_proto protoreflect.FileDescriptor

var file_internal_proto_workflow_v2_workflow_proto_rawDesc = []byte{
	0x0a, 0x29, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x76, 0x32, 0x2f, 0x77, 0x6f, 0x72,
	0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x22, 0x33, 0x0a, 0x16, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x19, 0x0a, 0x08, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x22, 0x5b, 0x0a, 0x17,
	0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66,
	0x6c, 0x6f, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66,
	0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x52,
	0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x22, 0x4e, 0x0a, 0x13, 0x50, 0x75, 0x62,
	0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x37, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x21, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x16, 0x0a, 0x14, 0x50, 0x75, 0x62,
	0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0xcc, 0x03, 0x0a, 0x08, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x12, 0x1f,
	0x0a, 0x0b, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x12,
	0x45, 0x0a, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x2b, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x57, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a, 0xd7, 0x02, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x15, 0x0a, 0x03, 0x63,
	0x6d, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x03, 0x63, 0x6d, 0x64, 0x88,
	0x01, 0x01, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x12, 0x46, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x06, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32,
	0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x45, 0x6e, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x65, 0x6e, 0x76, 0x12, 0x18,
	0x0a, 0x07, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x07, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x12, 0x30, 0x0a, 0x11, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x10, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x88, 0x01, 0x01, 0x1a, 0x36, 0x0a, 0x08, 0x45, 0x6e,
	0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x42, 0x06, 0x0a, 0x04, 0x5f, 0x63, 0x6d, 0x64, 0x42, 0x14, 0x0a, 0x12, 0x5f, 0x6e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x22, 0xcf, 0x04, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x12, 0x58, 0x0a, 0x0e, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32,
	0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61,
	0x72, 0x74, 0x65, 0x64, 0x48, 0x00, 0x52, 0x0d, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74,
	0x61, 0x72, 0x74, 0x65, 0x64, 0x12, 0x5e, 0x0a, 0x10, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x73, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x31, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64,
	0x65, 0x64, 0x48, 0x00, 0x52, 0x0f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x75, 0x63, 0x63,
	0x65, 0x65, 0x64, 0x65, 0x64, 0x12, 0x55, 0x0a, 0x0d, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e,
	0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x48, 0x00, 0x52, 0x0c,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x1a, 0x2c, 0x0a, 0x0d,
	0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x12, 0x1b, 0x0a,
	0x09, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x1a, 0x2e, 0x0a, 0x0f, 0x41, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x53, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x12, 0x1b, 0x0a,
	0x09, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x1a, 0xac, 0x01, 0x0a, 0x0c, 0x41,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x2a, 0x0a, 0x0e, 0x66, 0x61, 0x69, 0x6c,
	0x75, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x0d, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x65, 0x61, 0x73, 0x6f,
	0x6e, 0x88, 0x01, 0x01, 0x12, 0x2c, 0x0a, 0x0f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52,
	0x0e, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x88,
	0x01, 0x01, 0x42, 0x11, 0x0a, 0x0f, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x42, 0x12, 0x0a, 0x10, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72,
	0x65, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x07, 0x0a, 0x05, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x22, 0x11, 0x0a, 0x0f, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x86, 0x02, 0x0a, 0x0f, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7e, 0x0a, 0x0f, 0x53, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x12, 0x32, 0x2e, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x33, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x73, 0x0a, 0x0c, 0x50, 0x75, 0x62,
	0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x2f, 0x2e, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66,
	0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x30, 0x2e, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x76, 0x32, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x40,
	0x5a, 0x3e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x69, 0x6e,
	0x6b, 0x65, 0x72, 0x62, 0x65, 0x6c, 0x6c, 0x2f, 0x74, 0x69, 0x6e, 0x6b, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x76, 0x32, 0x3b, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_proto_workflow_v2_workflow_proto_rawDescOnce sync.Once
	file_internal_proto_workflow_v2_workflow_proto_rawDescData = file_internal_proto_workflow_v2_workflow_proto_rawDesc
)

func file_internal_proto_workflow_v2_workflow_proto_rawDescGZIP() []byte {
	file_internal_proto_workflow_v2_workflow_proto_rawDescOnce.Do(func() {
		file_internal_proto_workflow_v2_workflow_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_proto_workflow_v2_workflow_proto_rawDescData)
	})
	return file_internal_proto_workflow_v2_workflow_proto_rawDescData
}

var (
	file_internal_proto_workflow_v2_workflow_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
	file_internal_proto_workflow_v2_workflow_proto_goTypes  = []interface{}{
		(*StreamWorkflowsRequest)(nil),  // 0: internal.proto.workflow.v2.StreamWorkflowsRequest
		(*StreamWorkflowsResponse)(nil), // 1: internal.proto.workflow.v2.StreamWorkflowsResponse
		(*PublishEventRequest)(nil),     // 2: internal.proto.workflow.v2.PublishEventRequest
		(*PublishEventResponse)(nil),    // 3: internal.proto.workflow.v2.PublishEventResponse
		(*Workflow)(nil),                // 4: internal.proto.workflow.v2.Workflow
		(*Event)(nil),                   // 5: internal.proto.workflow.v2.Event
		(*PublishResponse)(nil),         // 6: internal.proto.workflow.v2.PublishResponse
		(*Workflow_Action)(nil),         // 7: internal.proto.workflow.v2.Workflow.Action
		nil,                             // 8: internal.proto.workflow.v2.Workflow.Action.EnvEntry
		(*Event_ActionStarted)(nil),     // 9: internal.proto.workflow.v2.Event.ActionStarted
		(*Event_ActionSucceeded)(nil),   // 10: internal.proto.workflow.v2.Event.ActionSucceeded
		(*Event_ActionFailed)(nil),      // 11: internal.proto.workflow.v2.Event.ActionFailed
	}
)
var file_internal_proto_workflow_v2_workflow_proto_depIdxs = []int32{
	4,  // 0: internal.proto.workflow.v2.StreamWorkflowsResponse.workflow:type_name -> internal.proto.workflow.v2.Workflow
	5,  // 1: internal.proto.workflow.v2.PublishEventRequest.event:type_name -> internal.proto.workflow.v2.Event
	7,  // 2: internal.proto.workflow.v2.Workflow.actions:type_name -> internal.proto.workflow.v2.Workflow.Action
	9,  // 3: internal.proto.workflow.v2.Event.action_started:type_name -> internal.proto.workflow.v2.Event.ActionStarted
	10, // 4: internal.proto.workflow.v2.Event.action_succeeded:type_name -> internal.proto.workflow.v2.Event.ActionSucceeded
	11, // 5: internal.proto.workflow.v2.Event.action_failed:type_name -> internal.proto.workflow.v2.Event.ActionFailed
	8,  // 6: internal.proto.workflow.v2.Workflow.Action.env:type_name -> internal.proto.workflow.v2.Workflow.Action.EnvEntry
	0,  // 7: internal.proto.workflow.v2.WorkflowService.StreamWorkflows:input_type -> internal.proto.workflow.v2.StreamWorkflowsRequest
	2,  // 8: internal.proto.workflow.v2.WorkflowService.PublishEvent:input_type -> internal.proto.workflow.v2.PublishEventRequest
	1,  // 9: internal.proto.workflow.v2.WorkflowService.StreamWorkflows:output_type -> internal.proto.workflow.v2.StreamWorkflowsResponse
	3,  // 10: internal.proto.workflow.v2.WorkflowService.PublishEvent:output_type -> internal.proto.workflow.v2.PublishEventResponse
	9,  // [9:11] is the sub-list for method output_type
	7,  // [7:9] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_internal_proto_workflow_v2_workflow_proto_init() }
func file_internal_proto_workflow_v2_workflow_proto_init() {
	if File_internal_proto_workflow_v2_workflow_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamWorkflowsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamWorkflowsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublishEventRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublishEventResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Workflow); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublishResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Workflow_Action); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event_ActionStarted); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event_ActionSucceeded); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_proto_workflow_v2_workflow_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event_ActionFailed); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_internal_proto_workflow_v2_workflow_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*Event_ActionStarted_)(nil),
		(*Event_ActionSucceeded_)(nil),
		(*Event_ActionFailed_)(nil),
	}
	file_internal_proto_workflow_v2_workflow_proto_msgTypes[7].OneofWrappers = []interface{}{}
	file_internal_proto_workflow_v2_workflow_proto_msgTypes[11].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_proto_workflow_v2_workflow_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_proto_workflow_v2_workflow_proto_goTypes,
		DependencyIndexes: file_internal_proto_workflow_v2_workflow_proto_depIdxs,
		MessageInfos:      file_internal_proto_workflow_v2_workflow_proto_msgTypes,
	}.Build()
	File_internal_proto_workflow_v2_workflow_proto = out.File
	file_internal_proto_workflow_v2_workflow_proto_rawDesc = nil
	file_internal_proto_workflow_v2_workflow_proto_goTypes = nil
	file_internal_proto_workflow_v2_workflow_proto_depIdxs = nil
}
