// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: meta.proto

/*
	Package meta is a generated protocol buffer package.

	It is generated from these files:
		meta.proto

	It has these top-level messages:
		Metadata
*/
package meta

import fmt "fmt"
import math "math"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.

// Metadata contains information about a file.
type Metadata struct {
	// the alignment size of the file.
	Size_ int `protobuf:"varint,1,opt,name=size,proto3,casttype=int" json:"size,omitempty"`
	// points at the first available record.
	Head int `protobuf:"varint,2,opt,name=head,proto3,casttype=int" json:"head,omitempty"`
	// advisory start and end of the records in the file.
	Start       int64 `protobuf:"varint,3,opt,name=start,proto3" json:"start,omitempty"`
	End         int64 `protobuf:"varint,4,opt,name=end,proto3" json:"end,omitempty"`
	SmallestEnd int64 `protobuf:"varint,5,opt,name=smallest_end,json=smallestEnd,proto3" json:"smallest_end,omitempty"`
}

func (m *Metadata) Reset()                    { *m = Metadata{} }
func (*Metadata) ProtoMessage()               {}
func (*Metadata) Descriptor() ([]byte, []int) { return fileDescriptorMeta, []int{0} }

func init() {
}
func (m *Metadata) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Metadata) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Size_ != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMeta(dAtA, i, uint64(m.Size_))
	}
	if m.Head != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintMeta(dAtA, i, uint64(m.Head))
	}
	if m.Start != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintMeta(dAtA, i, uint64(m.Start))
	}
	if m.End != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintMeta(dAtA, i, uint64(m.End))
	}
	if m.SmallestEnd != 0 {
		dAtA[i] = 0x28
		i++
		i = encodeVarintMeta(dAtA, i, uint64(m.SmallestEnd))
	}
	return i, nil
}

func encodeVarintMeta(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Metadata) Size() (n int) {
	var l int
	_ = l
	if m.Size_ != 0 {
		n += 1 + sovMeta(uint64(m.Size_))
	}
	if m.Head != 0 {
		n += 1 + sovMeta(uint64(m.Head))
	}
	if m.Start != 0 {
		n += 1 + sovMeta(uint64(m.Start))
	}
	if m.End != 0 {
		n += 1 + sovMeta(uint64(m.End))
	}
	if m.SmallestEnd != 0 {
		n += 1 + sovMeta(uint64(m.SmallestEnd))
	}
	return n
}

func sovMeta(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMeta(x uint64) (n int) {
	return sovMeta(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Metadata) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMeta
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Metadata: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Metadata: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Size_", wireType)
			}
			m.Size_ = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Size_ |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Head", wireType)
			}
			m.Head = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Head |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Start", wireType)
			}
			m.Start = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Start |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field End", wireType)
			}
			m.End = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.End |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SmallestEnd", wireType)
			}
			m.SmallestEnd = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SmallestEnd |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMeta(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMeta
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
func skipMeta(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMeta
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
					return 0, ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMeta
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthMeta
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMeta
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipMeta(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthMeta = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMeta   = fmt.Errorf("proto: integer overflow")
)


var fileDescriptorMeta = []byte{
	// 245 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x8f, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0x63, 0x92, 0x00, 0x32, 0x0c, 0xc8, 0x62, 0x88, 0x40, 0x32, 0x25, 0x53, 0x17, 0xdc,
	0x81, 0x37, 0xa8, 0xc4, 0xc8, 0xd2, 0x91, 0x05, 0x39, 0xf8, 0x9a, 0x58, 0x4d, 0x6c, 0x14, 0x5f,
	0x17, 0x5e, 0x82, 0xd7, 0xea, 0xc8, 0xc8, 0x84, 0xa8, 0x27, 0x46, 0x66, 0x26, 0xe4, 0x8b, 0x58,
	0xd8, 0xfe, 0xff, 0xfb, 0xec, 0xd3, 0x1d, 0xe7, 0x03, 0xa0, 0x56, 0xcf, 0xa3, 0x47, 0x2f, 0xea,
	0x30, 0xa8, 0xd1, 0x63, 0xb7, 0xf1, 0xca, 0xd8, 0xb0, 0x51, 0x6b, 0xdb, 0x43, 0x50, 0xd6, 0x21,
	0x8c, 0x4e, 0xf7, 0x2a, 0xbd, 0xbc, 0xb8, 0x69, 0x2d, 0x76, 0xdb, 0x46, 0x3d, 0xf9, 0x61, 0xd1,
	0xfa, 0xd6, 0x2f, 0xe8, 0x6b, 0xb3, 0x5d, 0x53, 0xa3, 0x42, 0x69, 0x1a, 0x59, 0xbf, 0x32, 0x7e,
	0x7c, 0x0f, 0xa8, 0x8d, 0x46, 0x2d, 0x2e, 0x79, 0x11, 0xec, 0x0b, 0x54, 0x6c, 0xc6, 0xe6, 0xe5,
	0xf2, 0xe8, 0xe7, 0xe3, 0x2a, 0xb7, 0x0e, 0x57, 0x04, 0x93, 0xec, 0x40, 0x9b, 0xea, 0xe0, 0x9f,
	0x4c, 0x50, 0x9c, 0xf3, 0x32, 0xa0, 0x1e, 0xb1, 0xca, 0x67, 0x6c, 0x9e, 0xaf, 0xa6, 0x22, 0xce,
	0x78, 0x0e, 0xce, 0x54, 0x05, 0xb1, 0x14, 0xc5, 0x35, 0x3f, 0x0d, 0x83, 0xee, 0x7b, 0x08, 0xf8,
	0x98, 0x54, 0x49, 0xea, 0xe4, 0x8f, 0xdd, 0x39, 0xb3, 0xac, 0x77, 0x7b, 0x99, 0xbd, 0xef, 0x65,
	0xb6, 0x8b, 0x92, 0xbd, 0x45, 0xc9, 0x3e, 0xa3, 0x64, 0x5f, 0x51, 0x66, 0xdf, 0x51, 0xb2, 0x87,
	0x22, 0x1d, 0xd9, 0x1c, 0xd2, 0xf2, 0xb7, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x11, 0x6a, 0xc8,
	0xa3, 0x1d, 0x01, 0x00, 0x00,
}
