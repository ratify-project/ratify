// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: referrerstore.proto

package referrerstore

import (
	common "github.com/deislabs/ratify/experimental/proto/v1/common"
	_struct "github.com/golang/protobuf/ptypes/struct"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The request object for ListReferrers.
type ListReferrersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The subject.
	Subject *common.Descriptor `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	// The specific types of artifacts to query and return. If empty, all artifacts regardless of type are returned.
	ArtifactTypes []string `protobuf:"bytes,2,rep,name=artifactTypes,proto3" json:"artifactTypes,omitempty"`
	// Optional. Custom to the store plugin. Can be used to modify the query performed or results returned, e.g: paging.
	Configuration *_struct.Struct `protobuf:"bytes,3,opt,name=configuration,proto3" json:"configuration,omitempty"`
}

func (x *ListReferrersRequest) Reset() {
	*x = ListReferrersRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_referrerstore_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListReferrersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListReferrersRequest) ProtoMessage() {}

func (x *ListReferrersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_referrerstore_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListReferrersRequest.ProtoReflect.Descriptor instead.
func (*ListReferrersRequest) Descriptor() ([]byte, []int) {
	return file_referrerstore_proto_rawDescGZIP(), []int{0}
}

func (x *ListReferrersRequest) GetSubject() *common.Descriptor {
	if x != nil {
		return x.Subject
	}
	return nil
}

func (x *ListReferrersRequest) GetArtifactTypes() []string {
	if x != nil {
		return x.ArtifactTypes
	}
	return nil
}

func (x *ListReferrersRequest) GetConfiguration() *_struct.Struct {
	if x != nil {
		return x.Configuration
	}
	return nil
}

// The response object for ListReferrers.
type ListReferrersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The subject.
	Subject *common.Descriptor `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	// The collection of results.
	Artifacts []*common.Referrer `protobuf:"bytes,2,rep,name=artifacts,proto3" json:"artifacts,omitempty"`
	// If paging is supported and more results were found,
	// this value can be provided in a follow up to get the next set.
	NextToken string `protobuf:"bytes,3,opt,name=nextToken,proto3" json:"nextToken,omitempty"`
}

func (x *ListReferrersResponse) Reset() {
	*x = ListReferrersResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_referrerstore_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListReferrersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListReferrersResponse) ProtoMessage() {}

func (x *ListReferrersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_referrerstore_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListReferrersResponse.ProtoReflect.Descriptor instead.
func (*ListReferrersResponse) Descriptor() ([]byte, []int) {
	return file_referrerstore_proto_rawDescGZIP(), []int{1}
}

func (x *ListReferrersResponse) GetSubject() *common.Descriptor {
	if x != nil {
		return x.Subject
	}
	return nil
}

func (x *ListReferrersResponse) GetArtifacts() []*common.Referrer {
	if x != nil {
		return x.Artifacts
	}
	return nil
}

func (x *ListReferrersResponse) GetNextToken() string {
	if x != nil {
		return x.NextToken
	}
	return ""
}

// The request for GetBlobContent.
type GetBlobContentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The artifact for which blobs have been requested.
	Artifact *common.Descriptor `protobuf:"bytes,1,opt,name=artifact,proto3" json:"artifact,omitempty"`
	// Optional. Custom to the store plugin. Can be used to modify the query performed or results returned.
	Configuration *_struct.Struct `protobuf:"bytes,2,opt,name=configuration,proto3" json:"configuration,omitempty"`
}

func (x *GetBlobContentRequest) Reset() {
	*x = GetBlobContentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_referrerstore_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlobContentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlobContentRequest) ProtoMessage() {}

func (x *GetBlobContentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_referrerstore_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlobContentRequest.ProtoReflect.Descriptor instead.
func (*GetBlobContentRequest) Descriptor() ([]byte, []int) {
	return file_referrerstore_proto_rawDescGZIP(), []int{2}
}

func (x *GetBlobContentRequest) GetArtifact() *common.Descriptor {
	if x != nil {
		return x.Artifact
	}
	return nil
}

func (x *GetBlobContentRequest) GetConfiguration() *_struct.Struct {
	if x != nil {
		return x.Configuration
	}
	return nil
}

// The response for GetBlobContent.
type GetBlobContentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The artifact for which blobs have been requested.
	Artifact *common.Descriptor `protobuf:"bytes,1,opt,name=artifact,proto3" json:"artifact,omitempty"`
	// The collection of blob contents.
	Content [][]byte `protobuf:"bytes,2,rep,name=content,proto3" json:"content,omitempty"`
}

func (x *GetBlobContentResponse) Reset() {
	*x = GetBlobContentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_referrerstore_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlobContentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlobContentResponse) ProtoMessage() {}

func (x *GetBlobContentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_referrerstore_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlobContentResponse.ProtoReflect.Descriptor instead.
func (*GetBlobContentResponse) Descriptor() ([]byte, []int) {
	return file_referrerstore_proto_rawDescGZIP(), []int{3}
}

func (x *GetBlobContentResponse) GetArtifact() *common.Descriptor {
	if x != nil {
		return x.Artifact
	}
	return nil
}

func (x *GetBlobContentResponse) GetContent() [][]byte {
	if x != nil {
		return x.Content
	}
	return nil
}

// The request for GetSubjectDescriptor.
type GetSubjectDescriptorRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The provided path for the subject.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Optional. Custom to the store plugin. Can be used to modify the query performed or results returned.
	Configuration *_struct.Struct `protobuf:"bytes,2,opt,name=configuration,proto3" json:"configuration,omitempty"`
}

func (x *GetSubjectDescriptorRequest) Reset() {
	*x = GetSubjectDescriptorRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_referrerstore_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetSubjectDescriptorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSubjectDescriptorRequest) ProtoMessage() {}

func (x *GetSubjectDescriptorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_referrerstore_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSubjectDescriptorRequest.ProtoReflect.Descriptor instead.
func (*GetSubjectDescriptorRequest) Descriptor() ([]byte, []int) {
	return file_referrerstore_proto_rawDescGZIP(), []int{4}
}

func (x *GetSubjectDescriptorRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *GetSubjectDescriptorRequest) GetConfiguration() *_struct.Struct {
	if x != nil {
		return x.Configuration
	}
	return nil
}

// The response for GetSubjectDescriptor.
type GetSubjectDescriptorResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The retrieved properties for the provided subject (path).
	Subject *common.Descriptor `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
}

func (x *GetSubjectDescriptorResponse) Reset() {
	*x = GetSubjectDescriptorResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_referrerstore_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetSubjectDescriptorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSubjectDescriptorResponse) ProtoMessage() {}

func (x *GetSubjectDescriptorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_referrerstore_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSubjectDescriptorResponse.ProtoReflect.Descriptor instead.
func (*GetSubjectDescriptorResponse) Descriptor() ([]byte, []int) {
	return file_referrerstore_proto_rawDescGZIP(), []int{5}
}

func (x *GetSubjectDescriptorResponse) GetSubject() *common.Descriptor {
	if x != nil {
		return x.Subject
	}
	return nil
}

// The request for GetReferenceManifest.
type GetManifestRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The path of the subject.
	SubjectPath string `protobuf:"bytes,1,opt,name=subjectPath,proto3" json:"subjectPath,omitempty"`
	// The referrer for which the manifest is being requested.
	Referrer *common.Descriptor `protobuf:"bytes,2,opt,name=referrer,proto3" json:"referrer,omitempty"`
	// Optional. Custom to the store plugin. Can be used to modify the query performed or results returned.
	Configuration *_struct.Struct `protobuf:"bytes,3,opt,name=configuration,proto3" json:"configuration,omitempty"`
}

func (x *GetManifestRequest) Reset() {
	*x = GetManifestRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_referrerstore_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetManifestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetManifestRequest) ProtoMessage() {}

func (x *GetManifestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_referrerstore_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetManifestRequest.ProtoReflect.Descriptor instead.
func (*GetManifestRequest) Descriptor() ([]byte, []int) {
	return file_referrerstore_proto_rawDescGZIP(), []int{6}
}

func (x *GetManifestRequest) GetSubjectPath() string {
	if x != nil {
		return x.SubjectPath
	}
	return ""
}

func (x *GetManifestRequest) GetReferrer() *common.Descriptor {
	if x != nil {
		return x.Referrer
	}
	return nil
}

func (x *GetManifestRequest) GetConfiguration() *_struct.Struct {
	if x != nil {
		return x.Configuration
	}
	return nil
}

// The response for GetReferenceManifest.
type GetManifestResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The fully realized Manifest object for the requested referrer.
	Manifest *common.Manifest `protobuf:"bytes,1,opt,name=manifest,proto3" json:"manifest,omitempty"`
}

func (x *GetManifestResponse) Reset() {
	*x = GetManifestResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_referrerstore_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetManifestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetManifestResponse) ProtoMessage() {}

func (x *GetManifestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_referrerstore_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetManifestResponse.ProtoReflect.Descriptor instead.
func (*GetManifestResponse) Descriptor() ([]byte, []int) {
	return file_referrerstore_proto_rawDescGZIP(), []int{7}
}

func (x *GetManifestResponse) GetManifest() *common.Manifest {
	if x != nil {
		return x.Manifest
	}
	return nil
}

var File_referrerstore_proto protoreflect.FileDescriptor

var file_referrerstore_proto_rawDesc = []byte{
	0x0a, 0x13, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xa9, 0x01, 0x0a, 0x14, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65,
	0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x07, 0x73, 0x75, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x07,
	0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x61, 0x72, 0x74, 0x69, 0x66,
	0x61, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0d,
	0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x73, 0x12, 0x3d, 0x0a,
	0x0d, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x0d, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x93, 0x01, 0x0a,
	0x15, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x07, 0x73, 0x75, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x12, 0x2e, 0x0a, 0x09, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x52, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x52, 0x09, 0x61, 0x72, 0x74, 0x69, 0x66,
	0x61, 0x63, 0x74, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x65, 0x78, 0x74, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x65, 0x78, 0x74, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x22, 0x86, 0x01, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a, 0x08,
	0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x6f, 0x72, 0x52, 0x08, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x12, 0x3d, 0x0a, 0x0d,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x0d, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x62, 0x0a, 0x16, 0x47,
	0x65, 0x74, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x08, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x08, 0x61, 0x72, 0x74,
	0x69, 0x66, 0x61, 0x63, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22,
	0x70, 0x0a, 0x1b, 0x47, 0x65, 0x74, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x44, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61,
	0x74, 0x68, 0x12, 0x3d, 0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x52, 0x0d, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x22, 0x4c, 0x0a, 0x1c, 0x47, 0x65, 0x74, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x44,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x2c, 0x0a, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x44, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x22,
	0xa5, 0x01, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x50, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x75, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x50, 0x61, 0x74, 0x68, 0x12, 0x2e, 0x0a, 0x08, 0x72, 0x65, 0x66, 0x65,
	0x72, 0x72, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x08,
	0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x12, 0x3d, 0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x0d, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x43, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x4d, 0x61,
	0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c,
	0x0a, 0x08, 0x6d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65,
	0x73, 0x74, 0x52, 0x08, 0x6d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x32, 0xa2, 0x03, 0x0a,
	0x13, 0x52, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x12, 0x5c, 0x0a, 0x0d, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x66, 0x65,
	0x72, 0x72, 0x65, 0x72, 0x73, 0x12, 0x23, 0x2e, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x72,
	0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x72, 0x65, 0x66,
	0x65, 0x72, 0x72, 0x65, 0x72, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52,
	0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x30, 0x01, 0x12, 0x5d, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x12, 0x24, 0x2e, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x72, 0x65, 0x66,
	0x65, 0x72, 0x72, 0x65, 0x72, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6c,
	0x6f, 0x62, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x6f, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x44,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x12, 0x2a, 0x2e, 0x72, 0x65, 0x66, 0x65,
	0x72, 0x72, 0x65, 0x72, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x75, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b, 0x2e, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x5d, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e,
	0x63, 0x65, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x12, 0x21, 0x2e, 0x72, 0x65, 0x66,
	0x65, 0x72, 0x72, 0x65, 0x72, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x4d, 0x61,
	0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e,
	0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x47, 0x65,
	0x74, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x40, 0x5a, 0x3e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x64, 0x65, 0x69, 0x73, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x72, 0x61, 0x74, 0x69, 0x66, 0x79, 0x2f,
	0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x73, 0x74,
	0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_referrerstore_proto_rawDescOnce sync.Once
	file_referrerstore_proto_rawDescData = file_referrerstore_proto_rawDesc
)

func file_referrerstore_proto_rawDescGZIP() []byte {
	file_referrerstore_proto_rawDescOnce.Do(func() {
		file_referrerstore_proto_rawDescData = protoimpl.X.CompressGZIP(file_referrerstore_proto_rawDescData)
	})
	return file_referrerstore_proto_rawDescData
}

var file_referrerstore_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_referrerstore_proto_goTypes = []interface{}{
	(*ListReferrersRequest)(nil),         // 0: referrerstore.ListReferrersRequest
	(*ListReferrersResponse)(nil),        // 1: referrerstore.ListReferrersResponse
	(*GetBlobContentRequest)(nil),        // 2: referrerstore.GetBlobContentRequest
	(*GetBlobContentResponse)(nil),       // 3: referrerstore.GetBlobContentResponse
	(*GetSubjectDescriptorRequest)(nil),  // 4: referrerstore.GetSubjectDescriptorRequest
	(*GetSubjectDescriptorResponse)(nil), // 5: referrerstore.GetSubjectDescriptorResponse
	(*GetManifestRequest)(nil),           // 6: referrerstore.GetManifestRequest
	(*GetManifestResponse)(nil),          // 7: referrerstore.GetManifestResponse
	(*common.Descriptor)(nil),            // 8: common.Descriptor
	(*_struct.Struct)(nil),               // 9: google.protobuf.Struct
	(*common.Referrer)(nil),              // 10: common.Referrer
	(*common.Manifest)(nil),              // 11: common.Manifest
}
var file_referrerstore_proto_depIdxs = []int32{
	8,  // 0: referrerstore.ListReferrersRequest.subject:type_name -> common.Descriptor
	9,  // 1: referrerstore.ListReferrersRequest.configuration:type_name -> google.protobuf.Struct
	8,  // 2: referrerstore.ListReferrersResponse.subject:type_name -> common.Descriptor
	10, // 3: referrerstore.ListReferrersResponse.artifacts:type_name -> common.Referrer
	8,  // 4: referrerstore.GetBlobContentRequest.artifact:type_name -> common.Descriptor
	9,  // 5: referrerstore.GetBlobContentRequest.configuration:type_name -> google.protobuf.Struct
	8,  // 6: referrerstore.GetBlobContentResponse.artifact:type_name -> common.Descriptor
	9,  // 7: referrerstore.GetSubjectDescriptorRequest.configuration:type_name -> google.protobuf.Struct
	8,  // 8: referrerstore.GetSubjectDescriptorResponse.subject:type_name -> common.Descriptor
	8,  // 9: referrerstore.GetManifestRequest.referrer:type_name -> common.Descriptor
	9,  // 10: referrerstore.GetManifestRequest.configuration:type_name -> google.protobuf.Struct
	11, // 11: referrerstore.GetManifestResponse.manifest:type_name -> common.Manifest
	0,  // 12: referrerstore.ReferrerStorePlugin.ListReferrers:input_type -> referrerstore.ListReferrersRequest
	2,  // 13: referrerstore.ReferrerStorePlugin.GetBlobContent:input_type -> referrerstore.GetBlobContentRequest
	4,  // 14: referrerstore.ReferrerStorePlugin.GetSubjectDescriptor:input_type -> referrerstore.GetSubjectDescriptorRequest
	6,  // 15: referrerstore.ReferrerStorePlugin.GetReferenceManifest:input_type -> referrerstore.GetManifestRequest
	1,  // 16: referrerstore.ReferrerStorePlugin.ListReferrers:output_type -> referrerstore.ListReferrersResponse
	3,  // 17: referrerstore.ReferrerStorePlugin.GetBlobContent:output_type -> referrerstore.GetBlobContentResponse
	5,  // 18: referrerstore.ReferrerStorePlugin.GetSubjectDescriptor:output_type -> referrerstore.GetSubjectDescriptorResponse
	7,  // 19: referrerstore.ReferrerStorePlugin.GetReferenceManifest:output_type -> referrerstore.GetManifestResponse
	16, // [16:20] is the sub-list for method output_type
	12, // [12:16] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_referrerstore_proto_init() }
func file_referrerstore_proto_init() {
	if File_referrerstore_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_referrerstore_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListReferrersRequest); i {
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
		file_referrerstore_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListReferrersResponse); i {
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
		file_referrerstore_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlobContentRequest); i {
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
		file_referrerstore_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlobContentResponse); i {
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
		file_referrerstore_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetSubjectDescriptorRequest); i {
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
		file_referrerstore_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetSubjectDescriptorResponse); i {
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
		file_referrerstore_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetManifestRequest); i {
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
		file_referrerstore_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetManifestResponse); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_referrerstore_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_referrerstore_proto_goTypes,
		DependencyIndexes: file_referrerstore_proto_depIdxs,
		MessageInfos:      file_referrerstore_proto_msgTypes,
	}.Build()
	File_referrerstore_proto = out.File
	file_referrerstore_proto_rawDesc = nil
	file_referrerstore_proto_goTypes = nil
	file_referrerstore_proto_depIdxs = nil
}
