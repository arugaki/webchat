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

package stringer

import (
	"code.google.com/p/gogoprotobuf/gogoproto"
	"code.google.com/p/gogoprotobuf/plugin/testgen"
	"code.google.com/p/gogoprotobuf/protoc-gen-gogo/generator"
)

type test struct {
	*generator.Generator
}

func NewTest(g *generator.Generator) testgen.TestPlugin {
	return &test{g}
}

func (p *test) Generate(imports generator.PluginImports, file *generator.FileDescriptor) bool {
	used := false
	randPkg := imports.NewImport("math/rand")
	timePkg := imports.NewImport("time")
	testingPkg := imports.NewImport("testing")
	fmtPkg := imports.NewImport("fmt")
	for _, message := range file.Messages() {
		ccTypeName := generator.CamelCaseSlice(message.TypeName())
		if !gogoproto.IsStringer(file.FileDescriptorProto, message.DescriptorProto) {
			continue
		}

		if gogoproto.HasTestGen(file.FileDescriptorProto, message.DescriptorProto) {
			used = true
			p.P(`func Test`, ccTypeName, `Stringer(t *`, testingPkg.Use(), `.T) {`)
			p.In()
			p.P(`popr := `, randPkg.Use(), `.New(`, randPkg.Use(), `.NewSource(`, timePkg.Use(), `.Now().UnixNano()))`)
			p.P(`p := NewPopulated`, ccTypeName, `(popr, false)`)
			p.P(`s1 := p.String()`)
			p.P(`s2 := `, fmtPkg.Use(), `.Sprintf("%v", p)`)
			p.P(`if s1 != s2 {`)
			p.In()
			p.P(`t.Fatalf("String want %v got %v", s1, s2)`)
			p.Out()
			p.P(`}`)
			p.Out()
			p.P(`}`)
		}

		if gogoproto.HasBenchGen(file.FileDescriptorProto, message.DescriptorProto) {
			used = true
			p.P(`func Benchmark`, ccTypeName, `Stringer(b *`, testingPkg.Use(), `.B) {`)
			p.In()
			p.P(`popr := `, randPkg.Use(), `.New(`, randPkg.Use(), `.NewSource(`, timePkg.Use(), `.Now().UnixNano()))`)
			p.P(`total := 0`)
			p.P(`b.ResetTimer()`)
			p.P(`b.StopTimer()`)
			p.P(`for i := 0; i < b.N; i++ {`)
			p.In()
			p.P(`p := NewPopulated`, ccTypeName, `(popr, true)`)
			p.P(`b.StartTimer()`)
			p.P(`data := p.String()`)
			p.P(`b.StopTimer()`)
			p.P(`total += len(data)`)
			p.Out()
			p.P(`}`)
			p.P(`b.SetBytes(int64(total / b.N))`)
			p.Out()
			p.P(`}`)
			p.P()
		}

	}
	return used
}

func init() {
	testgen.RegisterTestPlugin(NewTest)
}
