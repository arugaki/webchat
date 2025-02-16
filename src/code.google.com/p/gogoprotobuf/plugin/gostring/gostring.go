// Copyright (c) 2013, Vastech SA (PTY) LTD. All rights reserved.
// http://code.google.com/p/gogoprotobuf
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
The gostring plugin generates a GoString method for each message.
The GoString method is called whenever you use a fmt.Printf as such:

  fmt.Printf("%#v", mymessage)

or whenever you actually call GoString()
The output produced by the GoString method can be copied from the output into code and used to set a variable.
It is totally valid Go Code and is populated exactly as the struct that was printed out.

It is enabled by the following extensions:

  - gostring
  - gostring_all

The gostring plugin also generates a test given it is enabled using one of the following extensions:

  - testgen
  - testgen_all

And a benchmark given it is enabled using one of the following extensions:

  - benchgen
  - benchgen_all

Let us look at:

  code.google.com/p/gogoprotobuf/test/example/example.proto

Btw all the output can be seen at:

  code.google.com/p/gogoprotobuf/test/example/*

The following message:

  option (gogoproto.gostring_all) = true;

  message A {
	optional string Description = 1 [(gogoproto.nullable) = false];
	optional int64 Number = 2 [(gogoproto.nullable) = false];
	optional bytes Id = 3 [(gogoproto.customtype) = "code.google.com/p/gogoprotobuf/test/custom.Uuid", (gogoproto.nullable) = false];
  }

given to the gostring plugin, will generate the following code:

  func (this *A) GoString() string {
	if this == nil {
		return "nil"
	}
	s := strings1.Join([]string{`&test.A{` + `Description:` + fmt1.Sprintf("%#v", this.Description), `Number:` + fmt1.Sprintf("%#v", this.Number), `Id:` + fmt1.Sprintf("%#v", this.Id), `XXX_unrecognized:` + fmt1.Sprintf("%#v", this.XXX_unrecognized) + `}`}, ", ")
	return s
  }

and the following test code:

	func TestAGoString(t *testing6.T) {
		popr := math_rand6.New(math_rand6.NewSource(time6.Now().UnixNano()))
		p := NewPopulatedA(popr, false)
		s1 := p.GoString()
		s2 := fmt2.Sprintf("%#v", p)
		if s1 != s2 {
			t.Fatalf("GoString want %v got %v", s1, s2)
		}
		_, err := go_parser.ParseExpr(s1)
		if err != nil {
			panic(err)
		}
	}
	func BenchmarkAGoString(b *testing6.B) {
		popr := math_rand6.New(math_rand6.NewSource(time6.Now().UnixNano()))
		total := 0
		b.ResetTimer()
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			p := NewPopulatedA(popr, false)
			b.StartTimer()
			data := p.GoString()
			b.StopTimer()
			total += len(data)
		}
		b.SetBytes(int64(total / b.N))
	}

Typically fmt.Printf("%#v") will stop to print when it reaches a pointer and
not print their values, while the generated GoString method will always print all values, recursively.

*/
package gostring

import (
	"code.google.com/p/gogoprotobuf/gogoproto"
	"code.google.com/p/gogoprotobuf/protoc-gen-gogo/generator"
	"strings"
)

type gostring struct {
	*generator.Generator
	generator.PluginImports
	atleastOne bool
	localName  string
}

func NewGoString() *gostring {
	return &gostring{}
}

func (p *gostring) Name() string {
	return "gostring"
}

func (p *gostring) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *gostring) Generate(file *generator.FileDescriptor) {
	p.PluginImports = generator.NewPluginImports(p.Generator)
	p.atleastOne = false

	p.localName = generator.FileName(file)

	fmtPkg := p.NewImport("fmt")
	stringsPkg := p.NewImport("strings")
	protoPkg := p.NewImport("code.google.com/p/gogoprotobuf/proto")
	sortPkg := p.NewImport("sort")
	strconvPkg := p.NewImport("strconv")
	reflectPkg := p.NewImport("reflect")

	for _, message := range file.Messages() {
		if !gogoproto.HasGoString(file.FileDescriptorProto, message.DescriptorProto) {
			continue
		}
		p.atleastOne = true
		packageName := file.PackageName()

		ccTypeName := generator.CamelCaseSlice(message.TypeName())
		p.P(`func (this *`, ccTypeName, `) GoString() string {`)
		p.In()
		p.P(`if this == nil {`)
		p.In()
		p.P(`return "nil"`)
		p.Out()
		p.P(`}`)
		plus := "+"
		out := strings.Join([]string{"s := ", stringsPkg.Use(), ".Join([]string{`&", packageName, ".", ccTypeName, "{` ", plus, " "}, "")
		for _, field := range message.Field {
			nullable := gogoproto.IsNullable(field)
			repeated := field.IsRepeated()
			fieldname := generator.CamelCase(*field.Name)
			if field.IsMessage() || p.IsGroup(field) {
				desc := p.ObjectNamed(field.GetTypeName())
				msgname := p.TypeName(desc)
				msgnames := strings.Split(msgname, ".")
				typeName := msgnames[len(msgnames)-1]
				if gogoproto.IsEmbed(field) {
					fieldname = typeName
				}
				out += strings.Join([]string{"`", fieldname, ":` + "}, "")
				if nullable {
					out += strings.Join([]string{fmtPkg.Use(), `.Sprintf("%#v", this.`, fieldname, `)`}, "")
				} else if repeated {
					out += strings.Join([]string{stringsPkg.Use(), `.Replace(`, fmtPkg.Use(), `.Sprintf("%#v", this.`, fieldname, `)`, ",`&`,``,1)"}, "")
				} else {
					out += strings.Join([]string{stringsPkg.Use(), `.Replace(this.`, fieldname, `.GoString()`, ",`&`,``,1)"}, "")
				}
			} else {
				out += strings.Join([]string{"`", fieldname, ":` + "}, "")
				if field.IsEnum() {
					if nullable && !repeated {
						goTyp, _ := p.GoType(message, field)
						out += strings.Join([]string{`valueToGoString`, p.localName, `(this.`, fieldname, `,"`, packageName, ".", generator.GoTypeToName(goTyp), `"`, ")"}, "")
					} else {
						out += strings.Join([]string{fmtPkg.Use(), `.Sprintf("%#v", this.`, fieldname, ")"}, "")
					}
				} else {
					if nullable && !repeated {
						goTyp, _ := p.GoType(message, field)
						out += strings.Join([]string{`valueToGoString`, p.localName, `(this.`, fieldname, `,"`, generator.GoTypeToName(goTyp), `"`, ")"}, "")
					} else {
						out += strings.Join([]string{fmtPkg.Use(), `.Sprintf("%#v", this.`, fieldname, ")"}, "")
					}
				}
			}
			out += ", "
		}
		if message.DescriptorProto.HasExtension() {
			out += strings.Join([]string{"`XXX_extensions: ` + extensionToGoString", p.localName, `(this.XXX_extensions),`}, "")
		}
		out += strings.Join([]string{"`XXX_unrecognized:` + ", fmtPkg.Use(), `.Sprintf("%#v", this.XXX_unrecognized)`}, "")
		out += "+ `}`"
		out = strings.Join([]string{out, `}`, `,", "`, ")"}, "")
		p.P(out)
		p.P(`return s`)
		p.Out()
		p.P(`}`)
	}

	if !p.atleastOne {
		return
	}

	p.P(`func valueToGoString`, p.localName, `(v interface{}, typ string) string {`)
	p.In()
	p.P(`rv := `, reflectPkg.Use(), `.ValueOf(v)`)
	p.P(`if rv.IsNil() {`)
	p.In()
	p.P(`return "nil"`)
	p.Out()
	p.P(`}`)
	p.P(`pv := `, reflectPkg.Use(), `.Indirect(rv).Interface()`)
	p.P(`return `, fmtPkg.Use(), `.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)`)
	p.Out()
	p.P(`}`)

	p.P(`func extensionToGoString`, p.localName, `(e map[int32]`, protoPkg.Use(), `.Extension) string {`)
	p.In()
	p.P(`if e == nil { return "nil" }`)
	p.P(`s := "map[int32]proto.Extension{"`)
	p.P(`keys := make([]int, 0, len(e))`)
	p.P(`for k := range e {`)
	p.In()
	p.P(`keys = append(keys, int(k))`)
	p.Out()
	p.P(`}`)
	p.P(sortPkg.Use(), `.Ints(keys)`)
	p.P(`ss := []string{}`)
	p.P(`for _, k := range keys {`)
	p.In()
	p.P(`ss = append(ss, `, strconvPkg.Use(), `.Itoa(k) + ": " + e[int32(k)].GoString())`)
	p.Out()
	p.P(`}`)
	p.P(`s+=`, stringsPkg.Use(), `.Join(ss, ",") + "}"`)
	p.P(`return s`)
	p.Out()
	p.P(`}`)

}

func init() {
	generator.RegisterPlugin(NewGoString())
}
