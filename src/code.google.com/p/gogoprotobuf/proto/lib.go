// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2010 The Go Authors.  All rights reserved.
// http://code.google.com/p/goprotobuf/
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

/*
	Package proto converts data structures to and from the wire format of
	protocol buffers.  It works in concert with the Go source code generated
	for .proto files by the protocol compiler.

	A summary of the properties of the protocol buffer interface
	for a protocol buffer variable v:

	  - Names are turned from camel_case to CamelCase for export.
	  - There are no methods on v to set fields; just treat
	  	them as structure fields.
	  - There are getters that return a field's value if set,
		and return the field's default value if unset.
		The getters work even if the receiver is a nil message.
	  - The zero value for a struct is its correct initialization state.
		All desired fields must be set before marshaling.
	  - A Reset() method will restore a protobuf struct to its zero state.
	  - Non-repeated fields are pointers to the values; nil means unset.
		That is, optional or required field int32 f becomes F *int32.
	  - Repeated fields are slices.
	  - Helper functions are available to aid the setting of fields.
		Helpers for getting values are superseded by the
		GetFoo methods and their use is deprecated.
			msg.Foo = proto.String("hello") // set field
	  - Constants are defined to hold the default values of all fields that
		have them.  They have the form Default_StructName_FieldName.
		Because the getter methods handle defaulted values,
		direct use of these constants should be rare.
	  - Enums are given type names and maps from names to values.
		Enum values are prefixed with the enum's type name. Enum types have
		a String method, and a Enum method to assist in message construction.
	  - Nested groups and enums have type names prefixed with the name of
	  	the surrounding message type.
	  - Extensions are given descriptor names that start with E_,
		followed by an underscore-delimited list of the nested messages
		that contain it (if any) followed by the CamelCased name of the
		extension field itself.  HasExtension, ClearExtension, GetExtension
		and SetExtension are functions for manipulating extensions.
	  - Marshal and Unmarshal are functions to encode and decode the wire format.

	The simplest way to describe this is to see an example.
	Given file test.proto, containing

		package example;

		enum FOO { X = 17; };

		message Test {
		  required string label = 1;
		  optional int32 type = 2 [default=77];
		  repeated int64 reps = 3;
		  optional group OptionalGroup = 4 {
		    required string RequiredField = 5;
		  }
		}

	The resulting file, test.pb.go, is:

		package example

		import "code.google.com/p/gogoprotobuf/proto"

		type FOO int32
		const (
			FOO_X FOO = 17
		)
		var FOO_name = map[int32]string{
			17: "X",
		}
		var FOO_value = map[string]int32{
			"X": 17,
		}

		func (x FOO) Enum() *FOO {
			p := new(FOO)
			*p = x
			return p
		}
		func (x FOO) String() string {
			return proto.EnumName(FOO_name, int32(x))
		}

		type Test struct {
			Label            *string             `protobuf:"bytes,1,req,name=label" json:"label,omitempty"`
			Type             *int32              `protobuf:"varint,2,opt,name=type,def=77" json:"type,omitempty"`
			Reps             []int64             `protobuf:"varint,3,rep,name=reps" json:"reps,omitempty"`
			Optionalgroup    *Test_OptionalGroup `protobuf:"group,4,opt,name=OptionalGroup" json:"optionalgroup,omitempty"`
			XXX_unrecognized []byte              `json:"-"`
		}
		func (this *Test) Reset()         { *this = Test{} }
		func (this *Test) String() string { return proto.CompactTextString(this) }
		const Default_Test_Type int32 = 77

		func (this *Test) GetLabel() string {
			if this != nil && this.Label != nil {
				return *this.Label
			}
			return ""
		}

		func (this *Test) GetType() int32 {
			if this != nil && this.Type != nil {
				return *this.Type
			}
			return Default_Test_Type
		}

		func (this *Test) GetOptionalgroup() *Test_OptionalGroup {
			if this != nil {
				return this.Optionalgroup
			}
			return nil
		}

		type Test_OptionalGroup struct {
			RequiredField    *string `protobuf:"bytes,5,req" json:"RequiredField,omitempty"`
			XXX_unrecognized []byte  `json:"-"`
		}
		func (this *Test_OptionalGroup) Reset()         { *this = Test_OptionalGroup{} }
		func (this *Test_OptionalGroup) String() string { return proto.CompactTextString(this) }

		func (this *Test_OptionalGroup) GetRequiredField() string {
			if this != nil && this.RequiredField != nil {
				return *this.RequiredField
			}
			return ""
		}

		func init() {
			proto.RegisterEnum("example.FOO", FOO_name, FOO_value)
		}

	To create and play with a Test object:

		package main

		import (
			"log"

			"code.google.com/p/gogoprotobuf/proto"
			"./example.pb"
		)

		func main() {
			test := &example.Test{
				Label: proto.String("hello"),
				Type:  proto.Int32(17),
				Optionalgroup: &example.Test_OptionalGroup{
					RequiredField: proto.String("good bye"),
				},
			}
			data, err := proto.Marshal(test)
			if err != nil {
				log.Fatal("marshaling error: ", err)
			}
			newTest := new(example.Test)
			err = proto.Unmarshal(data, newTest)
			if err != nil {
				log.Fatal("unmarshaling error: ", err)
			}
			// Now test and newTest contain the same data.
			if test.GetLabel() != newTest.GetLabel() {
				log.Fatalf("data mismatch %q != %q", test.GetLabel(), newTest.GetLabel())
			}
			// etc.
		}
*/
package proto

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
)

// Message is implemented by generated protocol buffer messages.
type Message interface {
	Reset()
	String() string
	ProtoMessage()
}

// Stats records allocation details about the protocol buffer encoders
// and decoders.  Useful for tuning the library itself.
type Stats struct {
	Emalloc uint64 // mallocs in encode
	Dmalloc uint64 // mallocs in decode
	Encode  uint64 // number of encodes
	Decode  uint64 // number of decodes
	Chit    uint64 // number of cache hits
	Cmiss   uint64 // number of cache misses
}

// Set to true to enable stats collection.
const collectStats = false

var stats Stats

// GetStats returns a copy of the global Stats structure.
func GetStats() Stats { return stats }

// A Buffer is a buffer manager for marshaling and unmarshaling
// protocol buffers.  It may be reused between invocations to
// reduce memory usage.  It is not necessary to use a Buffer;
// the global functions Marshal and Unmarshal create a
// temporary Buffer and are fine for most applications.
type Buffer struct {
	buf       []byte     // encode/decode byte stream
	index     int        // write point
	freelist  [10][]byte // list of available buffers
	nfreelist int        // number of free buffers

	// pools of basic types to amortize allocation.
	bools   []bool
	uint32s []uint32
	uint64s []uint64

	// extra pools, only used with pointer_reflect.go
	int32s   []int32
	int64s   []int64
	float32s []float32
	float64s []float64
}

// NewBuffer allocates a new Buffer and initializes its internal data to
// the contents of the argument slice.
func NewBuffer(e []byte) *Buffer {
	p := new(Buffer)
	if e == nil {
		e = p.bufalloc()
	}
	p.buf = e
	p.index = 0
	return p
}

// Reset resets the Buffer, ready for marshaling a new protocol buffer.
func (p *Buffer) Reset() {
	if p.buf == nil {
		p.buf = p.bufalloc()
	}
	p.buf = p.buf[0:0] // for reading/writing
	p.index = 0        // for reading
}

// SetBuf replaces the internal buffer with the slice,
// ready for unmarshaling the contents of the slice.
func (p *Buffer) SetBuf(s []byte) {
	p.buf = s
	p.index = 0
}

// Bytes returns the contents of the Buffer.
func (p *Buffer) Bytes() []byte { return p.buf }

// Allocate a buffer for the Buffer.
func (p *Buffer) bufalloc() []byte {
	if p.nfreelist > 0 {
		// reuse an old one
		p.nfreelist--
		s := p.freelist[p.nfreelist]
		return s[0:0]
	}
	// make a new one
	s := make([]byte, 0, 16)
	return s
}

// Free (and remember in freelist) a byte buffer for the Buffer.
func (p *Buffer) buffree(s []byte) {
	if p.nfreelist < len(p.freelist) {
		// Take next slot.
		p.freelist[p.nfreelist] = s
		p.nfreelist++
		return
	}

	// Find the smallest.
	besti := -1
	bestl := len(s)
	for i, b := range p.freelist {
		if len(b) < bestl {
			besti = i
			bestl = len(b)
		}
	}

	// Overwrite the smallest.
	if besti >= 0 {
		p.freelist[besti] = s
	}
}

/*
 * Helper routines for simplifying the creation of optional fields of basic type.
 */

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// Int32 is a helper routine that allocates a new int32 value
// to store v and returns a pointer to it.
func Int32(v int32) *int32 {
	p := new(int32)
	*p = v
	return p
}

// Int is a helper routine that allocates a new int32 value
// to store v and returns a pointer to it, but unlike Int32
// its argument value is an int.
func Int(v int) *int32 {
	p := new(int32)
	*p = int32(v)
	return p
}

// Int64 is a helper routine that allocates a new int64 value
// to store v and returns a pointer to it.
func Int64(v int64) *int64 {
	p := new(int64)
	*p = v
	return p
}

// Float32 is a helper routine that allocates a new float32 value
// to store v and returns a pointer to it.
func Float32(v float32) *float32 {
	p := new(float32)
	*p = v
	return p
}

// Float64 is a helper routine that allocates a new float64 value
// to store v and returns a pointer to it.
func Float64(v float64) *float64 {
	p := new(float64)
	*p = v
	return p
}

// Uint32 is a helper routine that allocates a new uint32 value
// to store v and returns a pointer to it.
func Uint32(v uint32) *uint32 {
	p := new(uint32)
	*p = v
	return p
}

// Uint64 is a helper routine that allocates a new uint64 value
// to store v and returns a pointer to it.
func Uint64(v uint64) *uint64 {
	p := new(uint64)
	*p = v
	return p
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// EnumName is a helper function to simplify printing protocol buffer enums
// by name.  Given an enum map and a value, it returns a useful string.
func EnumName(m map[int32]string, v int32) string {
	s, ok := m[v]
	if ok {
		return s
	}
	return strconv.Itoa(int(v))
}

// UnmarshalJSONEnum is a helper function to simplify recovering enum int values
// from their JSON-encoded representation. Given a map from the enum's symbolic
// names to its int values, and a byte buffer containing the JSON-encoded
// value, it returns an int32 that can be cast to the enum type by the caller.
//
// The function can deal with older JSON representations, which represented
// enums directly by their int32 values, or with newer representations, which
// use the symbolic name as a string.
func UnmarshalJSONEnum(m map[string]int32, data []byte, enumName string) (int32, error) {
	if data[0] == '"' {
		// New style: enums are strings.
		var repr string
		if err := json.Unmarshal(data, &repr); err != nil {
			return -1, err
		}
		val, ok := m[repr]
		if !ok {
			return 0, fmt.Errorf("unrecognized enum %s value %q", enumName, repr)
		}
		return val, nil
	}
	// Old style: enums are ints.
	var val int32
	if err := json.Unmarshal(data, &val); err != nil {
		return 0, fmt.Errorf("cannot unmarshal %#q into enum %s", data, enumName)
	}
	return val, nil
}

// DebugPrint dumps the encoded data in b in a debugging format with a header
// including the string s. Used in testing but made available for general debugging.
func (o *Buffer) DebugPrint(s string, b []byte) {
	var u uint64

	obuf := o.buf
	index := o.index
	o.buf = b
	o.index = 0
	depth := 0

	fmt.Printf("\n--- %s ---\n", s)

out:
	for {
		for i := 0; i < depth; i++ {
			fmt.Print("  ")
		}

		index := o.index
		if index == len(o.buf) {
			break
		}

		op, err := o.DecodeVarint()
		if err != nil {
			fmt.Printf("%3d: fetching op err %v\n", index, err)
			break out
		}
		tag := op >> 3
		wire := op & 7

		switch wire {
		default:
			fmt.Printf("%3d: t=%3d unknown wire=%d\n",
				index, tag, wire)
			break out

		case WireBytes:
			var r []byte

			r, err = o.DecodeRawBytes(false)
			if err != nil {
				break out
			}
			fmt.Printf("%3d: t=%3d bytes [%d]", index, tag, len(r))
			if len(r) <= 6 {
				for i := 0; i < len(r); i++ {
					fmt.Printf(" %.2x", r[i])
				}
			} else {
				for i := 0; i < 3; i++ {
					fmt.Printf(" %.2x", r[i])
				}
				fmt.Printf(" ..")
				for i := len(r) - 3; i < len(r); i++ {
					fmt.Printf(" %.2x", r[i])
				}
			}
			fmt.Printf("\n")

		case WireFixed32:
			u, err = o.DecodeFixed32()
			if err != nil {
				fmt.Printf("%3d: t=%3d fix32 err %v\n", index, tag, err)
				break out
			}
			fmt.Printf("%3d: t=%3d fix32 %d\n", index, tag, u)

		case WireFixed64:
			u, err = o.DecodeFixed64()
			if err != nil {
				fmt.Printf("%3d: t=%3d fix64 err %v\n", index, tag, err)
				break out
			}
			fmt.Printf("%3d: t=%3d fix64 %d\n", index, tag, u)
			break

		case WireVarint:
			u, err = o.DecodeVarint()
			if err != nil {
				fmt.Printf("%3d: t=%3d varint err %v\n", index, tag, err)
				break out
			}
			fmt.Printf("%3d: t=%3d varint %d\n", index, tag, u)

		case WireStartGroup:
			if err != nil {
				fmt.Printf("%3d: t=%3d start err %v\n", index, tag, err)
				break out
			}
			fmt.Printf("%3d: t=%3d start\n", index, tag)
			depth++

		case WireEndGroup:
			depth--
			if err != nil {
				fmt.Printf("%3d: t=%3d end err %v\n", index, tag, err)
				break out
			}
			fmt.Printf("%3d: t=%3d end\n", index, tag)
		}
	}

	if depth != 0 {
		fmt.Printf("%3d: start-end not balanced %d\n", o.index, depth)
	}
	fmt.Printf("\n")

	o.buf = obuf
	o.index = index
}

// SetDefaults sets unset protocol buffer fields to their default values.
// It only modifies fields that are both unset and have defined defaults.
// It recursively sets default values in any non-nil sub-messages.
func SetDefaults(pb Message) {
	setDefaults(reflect.ValueOf(pb), true, false)
}

// v is a pointer to a struct.
func setDefaults(v reflect.Value, recur, zeros bool) {
	v = v.Elem()

	defaultMu.RLock()
	dm, ok := defaults[v.Type()]
	defaultMu.RUnlock()
	if !ok {
		dm = buildDefaultMessage(v.Type())
		defaultMu.Lock()
		defaults[v.Type()] = dm
		defaultMu.Unlock()
	}

	for _, sf := range dm.scalars {
		f := v.Field(sf.index)
		if !f.IsNil() {
			// field already set
			continue
		}
		dv := sf.value
		if dv == nil && !zeros {
			// no explicit default, and don't want to set zeros
			continue
		}
		fptr := f.Addr().Interface() // **T
		// TODO: Consider batching the allocations we do here.
		switch sf.kind {
		case reflect.Bool:
			b := new(bool)
			if dv != nil {
				*b = dv.(bool)
			}
			*(fptr.(**bool)) = b
		case reflect.Float32:
			f := new(float32)
			if dv != nil {
				*f = dv.(float32)
			}
			*(fptr.(**float32)) = f
		case reflect.Float64:
			f := new(float64)
			if dv != nil {
				*f = dv.(float64)
			}
			*(fptr.(**float64)) = f
		case reflect.Int32:
			// might be an enum
			if ft := f.Type(); ft != int32PtrType {
				// enum
				f.Set(reflect.New(ft.Elem()))
				if dv != nil {
					f.Elem().SetInt(int64(dv.(int32)))
				}
			} else {
				// int32 field
				i := new(int32)
				if dv != nil {
					*i = dv.(int32)
				}
				*(fptr.(**int32)) = i
			}
		case reflect.Int64:
			i := new(int64)
			if dv != nil {
				*i = dv.(int64)
			}
			*(fptr.(**int64)) = i
		case reflect.String:
			s := new(string)
			if dv != nil {
				*s = dv.(string)
			}
			*(fptr.(**string)) = s
		case reflect.Uint8:
			// exceptional case: []byte
			var b []byte
			if dv != nil {
				db := dv.([]byte)
				b = make([]byte, len(db))
				copy(b, db)
			} else {
				b = []byte{}
			}
			*(fptr.(*[]byte)) = b
		case reflect.Uint32:
			u := new(uint32)
			if dv != nil {
				*u = dv.(uint32)
			}
			*(fptr.(**uint32)) = u
		case reflect.Uint64:
			u := new(uint64)
			if dv != nil {
				*u = dv.(uint64)
			}
			*(fptr.(**uint64)) = u
		default:
			log.Printf("proto: can't set default for field %v (sf.kind=%v)", f, sf.kind)
		}
	}

	for _, ni := range dm.nested {
		f := v.Field(ni)
		if f.IsNil() {
			continue
		}
		setDefaults(f, recur, zeros)
	}
}

var (
	// defaults maps a protocol buffer struct type to a slice of the fields,
	// with its scalar fields set to their proto-declared non-zero default values.
	defaultMu sync.RWMutex
	defaults  = make(map[reflect.Type]defaultMessage)

	int32PtrType = reflect.TypeOf((*int32)(nil))
)

// defaultMessage represents information about the default values of a message.
type defaultMessage struct {
	scalars []scalarField
	nested  []int // struct field index of nested messages
}

type scalarField struct {
	index int          // struct field index
	kind  reflect.Kind // element type (the T in *T or []T)
	value interface{}  // the proto-declared default value, or nil
}

// t is a struct type.
func buildDefaultMessage(t reflect.Type) (dm defaultMessage) {
	sprop := GetProperties(t)
	for _, prop := range sprop.Prop {
		fi, ok := sprop.decoderTags.get(prop.Tag)
		if !ok {
			// XXX_unrecognized
			continue
		}
		ft := t.Field(fi).Type

		// nested messages
		if ft.Kind() == reflect.Ptr && ft.Elem().Kind() == reflect.Struct {
			dm.nested = append(dm.nested, fi)
			continue
		}

		sf := scalarField{
			index: fi,
			kind:  ft.Elem().Kind(),
		}

		// scalar fields without defaults
		if prop.Default == "" {
			dm.scalars = append(dm.scalars, sf)
			continue
		}

		// a scalar field: either *T or []byte
		switch ft.Elem().Kind() {
		case reflect.Bool:
			x, err := strconv.ParseBool(prop.Default)
			if err != nil {
				log.Printf("proto: bad default bool %q: %v", prop.Default, err)
				continue
			}
			sf.value = x
		case reflect.Float32:
			x, err := strconv.ParseFloat(prop.Default, 32)
			if err != nil {
				log.Printf("proto: bad default float32 %q: %v", prop.Default, err)
				continue
			}
			sf.value = float32(x)
		case reflect.Float64:
			x, err := strconv.ParseFloat(prop.Default, 64)
			if err != nil {
				log.Printf("proto: bad default float64 %q: %v", prop.Default, err)
				continue
			}
			sf.value = x
		case reflect.Int32:
			x, err := strconv.ParseInt(prop.Default, 10, 32)
			if err != nil {
				log.Printf("proto: bad default int32 %q: %v", prop.Default, err)
				continue
			}
			sf.value = int32(x)
		case reflect.Int64:
			x, err := strconv.ParseInt(prop.Default, 10, 64)
			if err != nil {
				log.Printf("proto: bad default int64 %q: %v", prop.Default, err)
				continue
			}
			sf.value = x
		case reflect.String:
			sf.value = prop.Default
		case reflect.Uint8:
			// []byte (not *uint8)
			sf.value = []byte(prop.Default)
		case reflect.Uint32:
			x, err := strconv.ParseUint(prop.Default, 10, 32)
			if err != nil {
				log.Printf("proto: bad default uint32 %q: %v", prop.Default, err)
				continue
			}
			sf.value = uint32(x)
		case reflect.Uint64:
			x, err := strconv.ParseUint(prop.Default, 10, 64)
			if err != nil {
				log.Printf("proto: bad default uint64 %q: %v", prop.Default, err)
				continue
			}
			sf.value = x
		default:
			log.Printf("proto: unhandled def kind %v", ft.Elem().Kind())
			continue
		}

		dm.scalars = append(dm.scalars, sf)
	}

	return dm
}
