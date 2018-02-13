package post

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Post) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Content":
			z.Content, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Date":
			z.Date, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "PubkeyStr":
			z.PubkeyStr, err = dc.ReadString()
			if err != nil {
				return
			}
		case "SigStr":
			z.SigStr, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Post) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "Content"
	err = en.Append(0x84, 0xa7, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = en.WriteString(z.Content)
	if err != nil {
		return
	}
	// write "Date"
	err = en.Append(0xa4, 0x44, 0x61, 0x74, 0x65)
	if err != nil {
		return
	}
	err = en.WriteTime(z.Date)
	if err != nil {
		return
	}
	// write "PubkeyStr"
	err = en.Append(0xa9, 0x50, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x53, 0x74, 0x72)
	if err != nil {
		return
	}
	err = en.WriteString(z.PubkeyStr)
	if err != nil {
		return
	}
	// write "SigStr"
	err = en.Append(0xa6, 0x53, 0x69, 0x67, 0x53, 0x74, 0x72)
	if err != nil {
		return
	}
	err = en.WriteString(z.SigStr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Post) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Content"
	o = append(o, 0x84, 0xa7, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74)
	o = msgp.AppendString(o, z.Content)
	// string "Date"
	o = append(o, 0xa4, 0x44, 0x61, 0x74, 0x65)
	o = msgp.AppendTime(o, z.Date)
	// string "PubkeyStr"
	o = append(o, 0xa9, 0x50, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x53, 0x74, 0x72)
	o = msgp.AppendString(o, z.PubkeyStr)
	// string "SigStr"
	o = append(o, 0xa6, 0x53, 0x69, 0x67, 0x53, 0x74, 0x72)
	o = msgp.AppendString(o, z.SigStr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Post) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Content":
			z.Content, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Date":
			z.Date, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "PubkeyStr":
			z.PubkeyStr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "SigStr":
			z.SigStr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Post) Msgsize() (s int) {
	s = 1 + 8 + msgp.StringPrefixSize + len(z.Content) + 5 + msgp.TimeSize + 10 + msgp.StringPrefixSize + len(z.PubkeyStr) + 7 + msgp.StringPrefixSize + len(z.SigStr)
	return
}
