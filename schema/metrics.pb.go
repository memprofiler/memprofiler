// Code generated by protoc-gen-go. DO NOT EDIT.
// source: metrics.proto

package schema

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// HeapConsumptionRates is a collection of rate values for memory consumption indicators.
// Formally, the rate (or velocity) is the first time derivative of any memory consumption indicator.
// For Bytes rate units are bytes per second, for Objects rate units are units per second
type HeapConsumptionRates struct {
	AllocObjects         float64  `protobuf:"fixed64,1,opt,name=AllocObjects,proto3" json:"AllocObjects,omitempty"`
	AllocBytes           float64  `protobuf:"fixed64,2,opt,name=AllocBytes,proto3" json:"AllocBytes,omitempty"`
	FreeObjects          float64  `protobuf:"fixed64,3,opt,name=FreeObjects,proto3" json:"FreeObjects,omitempty"`
	FreeBytes            float64  `protobuf:"fixed64,4,opt,name=FreeBytes,proto3" json:"FreeBytes,omitempty"`
	InUseObjects         float64  `protobuf:"fixed64,5,opt,name=InUseObjects,proto3" json:"InUseObjects,omitempty"`
	InUseBytes           float64  `protobuf:"fixed64,6,opt,name=InUseBytes,proto3" json:"InUseBytes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HeapConsumptionRates) Reset()         { *m = HeapConsumptionRates{} }
func (m *HeapConsumptionRates) String() string { return proto.CompactTextString(m) }
func (*HeapConsumptionRates) ProtoMessage()    {}
func (*HeapConsumptionRates) Descriptor() ([]byte, []int) {
	return fileDescriptor_6039342a2ba47b72, []int{0}
}

func (m *HeapConsumptionRates) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HeapConsumptionRates.Unmarshal(m, b)
}
func (m *HeapConsumptionRates) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HeapConsumptionRates.Marshal(b, m, deterministic)
}
func (m *HeapConsumptionRates) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HeapConsumptionRates.Merge(m, src)
}
func (m *HeapConsumptionRates) XXX_Size() int {
	return xxx_messageInfo_HeapConsumptionRates.Size(m)
}
func (m *HeapConsumptionRates) XXX_DiscardUnknown() {
	xxx_messageInfo_HeapConsumptionRates.DiscardUnknown(m)
}

var xxx_messageInfo_HeapConsumptionRates proto.InternalMessageInfo

func (m *HeapConsumptionRates) GetAllocObjects() float64 {
	if m != nil {
		return m.AllocObjects
	}
	return 0
}

func (m *HeapConsumptionRates) GetAllocBytes() float64 {
	if m != nil {
		return m.AllocBytes
	}
	return 0
}

func (m *HeapConsumptionRates) GetFreeObjects() float64 {
	if m != nil {
		return m.FreeObjects
	}
	return 0
}

func (m *HeapConsumptionRates) GetFreeBytes() float64 {
	if m != nil {
		return m.FreeBytes
	}
	return 0
}

func (m *HeapConsumptionRates) GetInUseObjects() float64 {
	if m != nil {
		return m.InUseObjects
	}
	return 0
}

func (m *HeapConsumptionRates) GetInUseBytes() float64 {
	if m != nil {
		return m.InUseBytes
	}
	return 0
}

// LocationMetrics is a set of heap allocation statistics
// that happened in a particular place in code
type LocationMetrics struct {
	// Rates represents heap consumption rates estimated
	// for some averaging window defined by server
	Rates *HeapConsumptionRates `protobuf:"bytes,1,opt,name=Rates,proto3" json:"Rates,omitempty"`
	// CallStack describes location in code where the allocation occured
	CallStack            *CallStack `protobuf:"bytes,2,opt,name=CallStack,proto3" json:"CallStack,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *LocationMetrics) Reset()         { *m = LocationMetrics{} }
func (m *LocationMetrics) String() string { return proto.CompactTextString(m) }
func (*LocationMetrics) ProtoMessage()    {}
func (*LocationMetrics) Descriptor() ([]byte, []int) {
	return fileDescriptor_6039342a2ba47b72, []int{1}
}

func (m *LocationMetrics) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocationMetrics.Unmarshal(m, b)
}
func (m *LocationMetrics) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocationMetrics.Marshal(b, m, deterministic)
}
func (m *LocationMetrics) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocationMetrics.Merge(m, src)
}
func (m *LocationMetrics) XXX_Size() int {
	return xxx_messageInfo_LocationMetrics.Size(m)
}
func (m *LocationMetrics) XXX_DiscardUnknown() {
	xxx_messageInfo_LocationMetrics.DiscardUnknown(m)
}

var xxx_messageInfo_LocationMetrics proto.InternalMessageInfo

func (m *LocationMetrics) GetRates() *HeapConsumptionRates {
	if m != nil {
		return m.Rates
	}
	return nil
}

func (m *LocationMetrics) GetCallStack() *CallStack {
	if m != nil {
		return m.CallStack
	}
	return nil
}

// SessionMetrics contains list of heap allocation metrics per every location
type SessionMetrics struct {
	Locations            []*LocationMetrics `protobuf:"bytes,1,rep,name=Locations,proto3" json:"Locations,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *SessionMetrics) Reset()         { *m = SessionMetrics{} }
func (m *SessionMetrics) String() string { return proto.CompactTextString(m) }
func (*SessionMetrics) ProtoMessage()    {}
func (*SessionMetrics) Descriptor() ([]byte, []int) {
	return fileDescriptor_6039342a2ba47b72, []int{2}
}

func (m *SessionMetrics) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionMetrics.Unmarshal(m, b)
}
func (m *SessionMetrics) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionMetrics.Marshal(b, m, deterministic)
}
func (m *SessionMetrics) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionMetrics.Merge(m, src)
}
func (m *SessionMetrics) XXX_Size() int {
	return xxx_messageInfo_SessionMetrics.Size(m)
}
func (m *SessionMetrics) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionMetrics.DiscardUnknown(m)
}

var xxx_messageInfo_SessionMetrics proto.InternalMessageInfo

func (m *SessionMetrics) GetLocations() []*LocationMetrics {
	if m != nil {
		return m.Locations
	}
	return nil
}

func init() {
	proto.RegisterType((*HeapConsumptionRates)(nil), "schema.HeapConsumptionRates")
	proto.RegisterType((*LocationMetrics)(nil), "schema.LocationMetrics")
	proto.RegisterType((*SessionMetrics)(nil), "schema.SessionMetrics")
}

func init() { proto.RegisterFile("metrics.proto", fileDescriptor_6039342a2ba47b72) }

var fileDescriptor_6039342a2ba47b72 = []byte{
	// 270 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x51, 0xed, 0x4a, 0x03, 0x31,
	0x10, 0xe4, 0xac, 0x3d, 0xb8, 0x3d, 0x3f, 0x68, 0x10, 0x2c, 0x52, 0xa4, 0xdc, 0xaf, 0xfe, 0x3a,
	0xe1, 0xc4, 0x07, 0xd0, 0x82, 0x1f, 0xa0, 0x08, 0x29, 0x3e, 0x40, 0x1a, 0x56, 0x3c, 0x4d, 0x2e,
	0x47, 0x12, 0x05, 0x5f, 0xd6, 0x67, 0x91, 0x6c, 0xb8, 0x0f, 0xa5, 0x3f, 0x77, 0x76, 0x32, 0x3b,
	0x33, 0x81, 0x43, 0x8d, 0xde, 0xd6, 0xd2, 0x95, 0xad, 0x35, 0xde, 0xb0, 0xd4, 0xc9, 0x37, 0xd4,
	0xe2, 0x6c, 0xa6, 0x51, 0xb7, 0xd6, 0xbc, 0xd6, 0x0a, 0x6d, 0x5c, 0x15, 0x3f, 0x09, 0x9c, 0xdc,
	0xa3, 0x68, 0xd7, 0xa6, 0x71, 0x9f, 0xba, 0xf5, 0xb5, 0x69, 0xb8, 0xf0, 0xe8, 0x58, 0x01, 0x07,
	0xd7, 0x4a, 0x19, 0xf9, 0xbc, 0x7d, 0x47, 0xe9, 0xdd, 0x3c, 0x59, 0x26, 0xab, 0x84, 0xff, 0xc1,
	0xd8, 0x39, 0x00, 0xcd, 0x37, 0xdf, 0x1e, 0xdd, 0x7c, 0x8f, 0x18, 0x23, 0x84, 0x2d, 0x21, 0xbf,
	0xb5, 0x88, 0x9d, 0xc4, 0x84, 0x08, 0x63, 0x88, 0x2d, 0x20, 0x0b, 0x63, 0x14, 0xd8, 0xa7, 0xfd,
	0x00, 0x04, 0x0f, 0x0f, 0xcd, 0x8b, 0xeb, 0x05, 0xa6, 0xd1, 0xc3, 0x18, 0x0b, 0x1e, 0x68, 0x8e,
	0x12, 0x69, 0xf4, 0x30, 0x20, 0xc5, 0x17, 0x1c, 0x3f, 0x1a, 0x29, 0x42, 0xb0, 0xa7, 0x58, 0x0a,
	0xab, 0x60, 0x4a, 0x19, 0x29, 0x53, 0x5e, 0x2d, 0xca, 0x58, 0x4f, 0xb9, 0xab, 0x07, 0x1e, 0xa9,
	0xec, 0x02, 0xb2, 0xb5, 0x50, 0x6a, 0xe3, 0x85, 0xfc, 0xa0, 0xa4, 0x79, 0x35, 0xeb, 0xde, 0xf5,
	0x0b, 0x3e, 0x70, 0x8a, 0x3b, 0x38, 0xda, 0xa0, 0x73, 0xa3, 0xb3, 0x57, 0x90, 0x75, 0x4e, 0xc2,
	0xe9, 0xc9, 0x2a, 0xaf, 0x4e, 0x3b, 0x89, 0x7f, 0x16, 0xf9, 0xc0, 0xdc, 0xa6, 0xf4, 0x51, 0x97,
	0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xf0, 0x86, 0x49, 0x3b, 0xd4, 0x01, 0x00, 0x00,
}
