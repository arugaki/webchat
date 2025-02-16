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
The embedcheck plugin is used to check whether embed is not used incorrectly.
For instance:
An embedded message has a generated string method, but the is a member of a message which does not.
This causes a warning.
An error is caused by a namespace conflict.

It is enabled by the following extensions:

  - embed
  - embed_all

For incorrect usage of embed with tests see:

  code.google.com/p/gogoprotobuf/test/embedconflict

*/
package embedcheck

import (
	"code.google.com/p/gogoprotobuf/gogoproto"
	"code.google.com/p/gogoprotobuf/protoc-gen-gogo/generator"
	"fmt"
	"os"
)

type plugin struct {
	*generator.Generator
}

func NewPlugin() *plugin {
	return &plugin{}
}

func (p *plugin) Name() string {
	return "embedcheck"
}

func (p *plugin) Init(g *generator.Generator) {
	p.Generator = g
}

var overwriters []map[string]gogoproto.EnableFunc = []map[string]gogoproto.EnableFunc{
	{
		"stringer": gogoproto.IsStringer,
	},
	{
		"gostring": gogoproto.HasGoString,
	},
	{
		"equal": gogoproto.HasEqual,
	},
	{
		"verboseequal": gogoproto.HasVerboseEqual,
	},
	{
		"size": gogoproto.IsSizer,
	},
	{
		"unmarshaler":        gogoproto.IsUnmarshaler,
		"unsafe_unmarshaler": gogoproto.IsUnsafeUnmarshaler,
	},
	{
		"marshaler":        gogoproto.IsMarshaler,
		"unsafe_marshaler": gogoproto.IsUnsafeMarshaler,
	},
}

func (p *plugin) Generate(file *generator.FileDescriptor) {
	for _, msg := range file.Messages() {
		for _, os := range overwriters {
			possible := true
			for _, overwriter := range os {
				if overwriter(file.FileDescriptorProto, msg.DescriptorProto) {
					possible = false
				}
			}
			if possible {
				p.checkOverwrite(msg, os)
			}
		}
		p.checkNameSpace(msg)
	}
}

func (p *plugin) checkNameSpace(message *generator.Descriptor) map[string]bool {
	ccTypeName := generator.CamelCaseSlice(message.TypeName())
	names := make(map[string]bool)
	for _, field := range message.Field {
		fieldname := generator.CamelCase(*field.Name)
		if gogoproto.IsEmbed(field) {
			desc := p.ObjectNamed(field.GetTypeName())
			moreNames := p.checkNameSpace(desc.(*generator.Descriptor))
			for another := range moreNames {
				if names[another] {
					fmt.Fprintf(os.Stderr, "ERROR: duplicate embedded fieldname %v in type %v\n", fieldname, ccTypeName)
					os.Exit(1)
				}
				names[another] = true
			}
		} else {
			if names[fieldname] {
				fmt.Fprintf(os.Stderr, "ERROR: duplicate embedded fieldname %v in type %v\n", fieldname, ccTypeName)
				os.Exit(1)
			}
			names[fieldname] = true
		}
	}
	return names
}

func (p *plugin) checkOverwrite(message *generator.Descriptor, enablers map[string]gogoproto.EnableFunc) {
	ccTypeName := generator.CamelCaseSlice(message.TypeName())
	names := []string{}
	for name := range enablers {
		names = append(names, name)
	}
	for _, field := range message.Field {
		if gogoproto.IsEmbed(field) {
			fieldname := generator.CamelCase(*field.Name)
			desc := p.ObjectNamed(field.GetTypeName())
			msg := desc.(*generator.Descriptor)
			for errStr, enabled := range enablers {
				if enabled(msg.File(), msg.DescriptorProto) {
					fmt.Fprintf(os.Stderr, "WARNING: found non-%v %v with embedded %v %v\n", names, ccTypeName, errStr, fieldname)
				}
			}
			p.checkOverwrite(msg, enablers)
		}
	}
}

func (p *plugin) GenerateImports(*generator.FileDescriptor) {}

func init() {
	generator.RegisterPlugin(NewPlugin())
}
