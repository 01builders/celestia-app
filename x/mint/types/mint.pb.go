// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: celestia/mint/v1/mint.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Minter represents the mint state.
type Minter struct {
	// InflationRate is the rate at which new tokens should be minted for the
	// current year. For example if InflationRate=0.1, then 10% of the total
	// supply will be minted over the course of the year.
	InflationRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,1,opt,name=inflation_rate,json=inflationRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"inflation_rate"`
	// AnnualProvisions is the total number of tokens to be minted in the current
	// year due to inflation.
	AnnualProvisions github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,2,opt,name=annual_provisions,json=annualProvisions,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"annual_provisions"`
	// PreviousBlockTime is the timestamp of the previous block.
	PreviousBlockTime *time.Time `protobuf:"bytes,4,opt,name=previous_block_time,json=previousBlockTime,proto3,stdtime" json:"previous_block_time,omitempty"`
	// BondDenom is the denomination of the token that should be minted.
	BondDenom string `protobuf:"bytes,5,opt,name=bond_denom,json=bondDenom,proto3" json:"bond_denom,omitempty"`
}

func (m *Minter) Reset()         { *m = Minter{} }
func (m *Minter) String() string { return proto.CompactTextString(m) }
func (*Minter) ProtoMessage()    {}
func (*Minter) Descriptor() ([]byte, []int) {
	return fileDescriptor_962d7cf1c9c59571, []int{0}
}
func (m *Minter) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Minter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Minter.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Minter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Minter.Merge(m, src)
}
func (m *Minter) XXX_Size() int {
	return m.Size()
}
func (m *Minter) XXX_DiscardUnknown() {
	xxx_messageInfo_Minter.DiscardUnknown(m)
}

var xxx_messageInfo_Minter proto.InternalMessageInfo

func (m *Minter) GetPreviousBlockTime() *time.Time {
	if m != nil {
		return m.PreviousBlockTime
	}
	return nil
}

func (m *Minter) GetBondDenom() string {
	if m != nil {
		return m.BondDenom
	}
	return ""
}

// GenesisTime contains the timestamp of the genesis block.
type GenesisTime struct {
	// GenesisTime is the timestamp of the genesis block.
	GenesisTime *time.Time `protobuf:"bytes,1,opt,name=genesis_time,json=genesisTime,proto3,stdtime" json:"genesis_time,omitempty"`
}

func (m *GenesisTime) Reset()         { *m = GenesisTime{} }
func (m *GenesisTime) String() string { return proto.CompactTextString(m) }
func (*GenesisTime) ProtoMessage()    {}
func (*GenesisTime) Descriptor() ([]byte, []int) {
	return fileDescriptor_962d7cf1c9c59571, []int{1}
}
func (m *GenesisTime) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisTime) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisTime.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisTime) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisTime.Merge(m, src)
}
func (m *GenesisTime) XXX_Size() int {
	return m.Size()
}
func (m *GenesisTime) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisTime.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisTime proto.InternalMessageInfo

func (m *GenesisTime) GetGenesisTime() *time.Time {
	if m != nil {
		return m.GenesisTime
	}
	return nil
}

func init() {
	proto.RegisterType((*Minter)(nil), "celestia.mint.v1.Minter")
	proto.RegisterType((*GenesisTime)(nil), "celestia.mint.v1.GenesisTime")
}

func init() { proto.RegisterFile("celestia/mint/v1/mint.proto", fileDescriptor_962d7cf1c9c59571) }

var fileDescriptor_962d7cf1c9c59571 = []byte{
	// 383 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x52, 0xc1, 0xca, 0xd3, 0x40,
	0x10, 0xce, 0x96, 0x5a, 0xe8, 0x56, 0xa5, 0x8d, 0x1e, 0x62, 0xc5, 0xa4, 0xf4, 0x20, 0xbd, 0x34,
	0xb1, 0x7a, 0xf5, 0x14, 0x0b, 0x82, 0x20, 0x94, 0xe0, 0xc9, 0x4b, 0xd8, 0xa4, 0xdb, 0x75, 0x69,
	0xb2, 0x13, 0xb2, 0x9b, 0xa0, 0x6f, 0xd1, 0x87, 0xf1, 0x21, 0xea, 0xad, 0x78, 0x12, 0x0f, 0x55,
	0xda, 0x17, 0x91, 0xcd, 0x26, 0xd5, 0xa3, 0x87, 0xff, 0xb4, 0x33, 0xf3, 0xcd, 0x7c, 0xdf, 0xec,
	0xc7, 0xe0, 0xa7, 0x29, 0xcd, 0xa8, 0x54, 0x9c, 0x04, 0x39, 0x17, 0x2a, 0xa8, 0x57, 0xcd, 0xeb,
	0x17, 0x25, 0x28, 0xb0, 0xc7, 0x1d, 0xe8, 0x37, 0xc5, 0x7a, 0x35, 0x7d, 0xcc, 0x80, 0x41, 0x03,
	0x06, 0x3a, 0x32, 0x7d, 0xd3, 0x27, 0x29, 0xc8, 0x1c, 0x64, 0x6c, 0x00, 0x93, 0xb4, 0x90, 0xc7,
	0x00, 0x58, 0x46, 0x83, 0x26, 0x4b, 0xaa, 0x5d, 0xa0, 0x78, 0x4e, 0xa5, 0x22, 0x79, 0x61, 0x1a,
	0xe6, 0xdf, 0x7a, 0x78, 0xf0, 0x9e, 0x0b, 0x45, 0x4b, 0x3b, 0xc5, 0x0f, 0xb9, 0xd8, 0x65, 0x44,
	0x71, 0x10, 0x71, 0x49, 0x14, 0x75, 0xd0, 0x0c, 0x2d, 0x86, 0xe1, 0xeb, 0xe3, 0xd9, 0xb3, 0x7e,
	0x9e, 0xbd, 0xe7, 0x8c, 0xab, 0x4f, 0x55, 0xe2, 0xa7, 0x90, 0xb7, 0x22, 0xed, 0xb3, 0x94, 0xdb,
	0x7d, 0xa0, 0xbe, 0x14, 0x54, 0xfa, 0x6b, 0x9a, 0x7e, 0xff, 0xba, 0xc4, 0xed, 0x0e, 0x6b, 0x9a,
	0x46, 0x0f, 0x6e, 0x9c, 0x11, 0x51, 0xd4, 0xe6, 0x78, 0x42, 0x84, 0xa8, 0x48, 0xa6, 0xb7, 0xad,
	0xb9, 0xe4, 0x20, 0xa4, 0xd3, 0xbb, 0x03, 0x9d, 0xb1, 0xa1, 0xdd, 0xdc, 0x58, 0xed, 0x0d, 0x7e,
	0x54, 0x94, 0xb4, 0xe6, 0x50, 0xc9, 0x38, 0xc9, 0x20, 0xdd, 0xc7, 0xfa, 0xf3, 0x4e, 0x7f, 0x86,
	0x16, 0xa3, 0x97, 0x53, 0xdf, 0x38, 0xe3, 0x77, 0xce, 0xf8, 0x1f, 0x3a, 0x67, 0xc2, 0xfe, 0xe1,
	0x97, 0x87, 0xa2, 0x49, 0x37, 0x1c, 0xea, 0x59, 0x8d, 0xda, 0xcf, 0x30, 0x4e, 0x40, 0x6c, 0xe3,
	0x2d, 0x15, 0x90, 0x3b, 0xf7, 0xf4, 0xd6, 0xd1, 0x50, 0x57, 0xd6, 0xba, 0x30, 0x8f, 0xf0, 0xe8,
	0x2d, 0x15, 0x54, 0x72, 0xd9, 0x74, 0xbf, 0xc1, 0xf7, 0x99, 0x49, 0x8d, 0x30, 0xfa, 0x4f, 0xe1,
	0x11, 0xfb, 0x4b, 0x12, 0xbe, 0x3b, 0x5e, 0x5c, 0x74, 0xba, 0xb8, 0xe8, 0xf7, 0xc5, 0x45, 0x87,
	0xab, 0x6b, 0x9d, 0xae, 0xae, 0xf5, 0xe3, 0xea, 0x5a, 0x1f, 0x5f, 0xfc, 0x6b, 0x53, 0x7b, 0x28,
	0x50, 0xb2, 0x5b, 0xbc, 0x24, 0x45, 0x11, 0x7c, 0x36, 0x77, 0xd5, 0x98, 0x96, 0x0c, 0x1a, 0xc9,
	0x57, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x71, 0xfd, 0xfc, 0xbb, 0x75, 0x02, 0x00, 0x00,
}

func (m *Minter) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Minter) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Minter) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.BondDenom) > 0 {
		i -= len(m.BondDenom)
		copy(dAtA[i:], m.BondDenom)
		i = encodeVarintMint(dAtA, i, uint64(len(m.BondDenom)))
		i--
		dAtA[i] = 0x2a
	}
	if m.PreviousBlockTime != nil {
		n1, err1 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(*m.PreviousBlockTime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(*m.PreviousBlockTime):])
		if err1 != nil {
			return 0, err1
		}
		i -= n1
		i = encodeVarintMint(dAtA, i, uint64(n1))
		i--
		dAtA[i] = 0x22
	}
	{
		size := m.AnnualProvisions.Size()
		i -= size
		if _, err := m.AnnualProvisions.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintMint(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size := m.InflationRate.Size()
		i -= size
		if _, err := m.InflationRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintMint(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *GenesisTime) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisTime) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisTime) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.GenesisTime != nil {
		n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(*m.GenesisTime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(*m.GenesisTime):])
		if err2 != nil {
			return 0, err2
		}
		i -= n2
		i = encodeVarintMint(dAtA, i, uint64(n2))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintMint(dAtA []byte, offset int, v uint64) int {
	offset -= sovMint(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Minter) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.InflationRate.Size()
	n += 1 + l + sovMint(uint64(l))
	l = m.AnnualProvisions.Size()
	n += 1 + l + sovMint(uint64(l))
	if m.PreviousBlockTime != nil {
		l = github_com_cosmos_gogoproto_types.SizeOfStdTime(*m.PreviousBlockTime)
		n += 1 + l + sovMint(uint64(l))
	}
	l = len(m.BondDenom)
	if l > 0 {
		n += 1 + l + sovMint(uint64(l))
	}
	return n
}

func (m *GenesisTime) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.GenesisTime != nil {
		l = github_com_cosmos_gogoproto_types.SizeOfStdTime(*m.GenesisTime)
		n += 1 + l + sovMint(uint64(l))
	}
	return n
}

func sovMint(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMint(x uint64) (n int) {
	return sovMint(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Minter) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMint
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Minter: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Minter: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InflationRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.InflationRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AnnualProvisions", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.AnnualProvisions.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PreviousBlockTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.PreviousBlockTime == nil {
				m.PreviousBlockTime = new(time.Time)
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(m.PreviousBlockTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BondDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BondDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMint
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GenesisTime) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMint
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GenesisTime: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisTime: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GenesisTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.GenesisTime == nil {
				m.GenesisTime = new(time.Time)
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(m.GenesisTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMint
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMint(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMint
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMint
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMint
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthMint
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMint
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMint
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMint        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMint          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMint = fmt.Errorf("proto: unexpected end of group")
)
