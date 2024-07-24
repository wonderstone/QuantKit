// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

// gen-benchmarks generates benchmark tests based on the implemented functions.
//
// It is meant to be used by the authors to update the benchmark tests.
package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"sort"
	"text/template"

	in "github.com/wonderstone/QuantKit/modelgene/gep/functions/int_nodes"
)

const (
	filename = "ints-benchmarks_test.go"
	pkgName  = "intNodes"
)

var (
	verbose = flag.Bool("v", false, "Print verbose log messages")

	funcMap = map[string]interface{}{"validateFunc": validateFunc}

	sourceTmpl = template.Must(template.New("source").Funcs(funcMap).Parse(source))
)

func main() {
	flag.Parse()

	var funcs []string
	for k := range in.Int {
		funcs = append(funcs, k)
	}

	t := &templateData{
		Package: pkgName,
		Funcs:   funcs,
	}
	if err := t.dump(); err != nil {
		log.Fatal(err)
	}

	logf("Done.")
}

func validateFunc(in string) string {
	switch in {
	case "+":
		return "Plus"
	case "-":
		return "Minus"
	case "*":
		return "Multiply"
	case "/":
		return "Divide"
	default:
		return in
	}
}

func logf(fmt string, args ...interface{}) {
	if *verbose {
		log.Printf(fmt, args...)
	}
}

func (t *templateData) dump() error {
	if len(t.Funcs) == 0 {
		logf("No funcs for %v; skipping.", filename)
		return nil
	}

	// Sort funcs by ReceiverType.FieldName.
	sort.Strings(t.Funcs)

	var buf bytes.Buffer
	if err := sourceTmpl.Execute(&buf, t); err != nil {
		return err
	}
	clean, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	logf("Writing %v...", filename)
	return ioutil.WriteFile(filename, clean, 0644)
}

type templateData struct {
	Package string
	Funcs   []string
}

const source = `// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Code generated by gen-benchmarks; DO NOT EDIT.

package {{.Package}}

import (
  "testing"
)
{{range .Funcs}}
func Benchmark{{. | validateFunc}}(b *testing.B) { runBenchmark(b, "{{.}}") }
{{- end}}
`
