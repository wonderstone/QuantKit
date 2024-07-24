// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package genome

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go/format"

	"github.com/wonderstone/QuantKit/modelgene/gep/functions"
	"github.com/wonderstone/QuantKit/modelgene/gep/grammars"
)

type dump struct {
	w      io.Writer
	gr     *grammars.Grammar
	fm     functions.FuncMap
	genome *Genome
	subs   map[string]string
}

// this method in master branch is not finished yet ,if it can be done for complicated math_nodes
// And link function is only applied to its ArgOrder genes, a little bit messy
// use WriteExps, until the master branch is finished
func (g *Genome) Write(w io.Writer, grammar *grammars.Grammar) {
	d := &dump{
		gr:     grammar,
		genome: g,
		subs: map[string]string{
			"CHARX": "X",
		},
	}
	code, err := d.generateCode()
	if err != nil {
		fmt.Printf("error generating code: %v", err)
	}
	fmt.Fprintf(w, "%s", code)
}

func (g *Genome) WriteExps(w io.Writer, grammar *grammars.Grammar, indiNames []string) {
	// fmt.Println("+++++++++++++++++++++++++++")
	// fmt.Println(grammar, g, indiNames)
	d := &dump{
		gr:     grammar,
		genome: g,
		subs: map[string]string{
			"CHARX": "X",
		},
	}

	indiNamesMap := make(map[string]string)
	for i, n := range indiNames {
		indiNamesMap[fmt.Sprintf("d[%d]", i)] = n
	}
	// fmt.Println("************************************")
	// fmt.Println(indiNamesMap)
	// fmt.Println(d)
	exps, err := d.generateExps(indiNamesMap)
	if err != nil {
		fmt.Printf("error generating exps: %v", err)
	}
	fmt.Fprintf(w, "%s", exps)
}

// this part is for test only comment when finished the testing
func (g *Genome) WriteExpsStr(grammar *grammars.Grammar, indiNames []string) (res string) {
	d := &dump{
		gr:     grammar,
		genome: g,
		subs: map[string]string{
			"CHARX": "X",
		},
	}

	indiNamesMap := make(map[string]string)
	for i, n := range indiNames {
		indiNamesMap[fmt.Sprintf("d%d", i)] = n
	}
	// fmt.Println("************************************")
	// fmt.Println(d)
	exps, err := d.generateExps(indiNamesMap)
	if err != nil {
		fmt.Printf("error generating exps: %v", err)
	}
	// output the exps to a string
	return string(exps)
}

// for realtime job, output the simplified genome
func (g *Genome) SimplifyGenome(grammar *grammars.Grammar, indiNames []string) ([]string, error) {
	d := &dump{
		gr:     grammar,
		genome: g,
		subs: map[string]string{
			"CHARX": "X",
		},
	}
	indiNmsMap := make(map[string]string)            //key: d[i]  val: indicatorname i is the old index
	indiNmsMapSliceUsed := make(map[string][]string) // key:d[i] val: [dj, indicatorname] j is the new index
	mapping := make(map[string]string)               // key: di val:d[i] only for used indicators
	indiNmSlice := make([]string, 0)                 // [indicatorname,...,] order as j
	for i, n := range indiNames {
		indiNmsMap[fmt.Sprintf("d[%d]", i)] = n
	}
	// Generate the expression, keeping track of any helper functions that are needed.
	helpers := make(grammars.HelperMap)
	// iter the gene  in d and get the expression
	var newindex int = 0
	for _, e := range d.genome.Genes {
		exp, err := e.Expression(d.gr, helpers)
		if err != nil {
			return nil, err
		}
		for k, v := range indiNmsMap { // k: d[i] v: indicatorname
			if strings.Contains(exp, k) {
				// if k is not in indiNmsUsed
				if _, ok := indiNmsMapSliceUsed[k]; !ok {
					indiNmsMapSliceUsed[k] = []string{fmt.Sprintf("d%d", newindex), v}
					// remove "[" and "]" strings in k
					tmpk := k
					tmpk = strings.Replace(tmpk, "[", "", -1)
					tmpk = strings.Replace(tmpk, "]", "", -1)
					mapping[fmt.Sprintf(tmpk)] = k // bug you fixed
					indiNmSlice = append(indiNmSlice, v)
					newindex = newindex + 1
				}
			}
		}
	}

	// adjust the g *Genome
	re4terminals := regexp.MustCompile("d[0-9]+")
	for i, gene := range g.Genes {
		// iter the Symbols, change the d[i] which is not in the indiNmsUsed to the element in it
		for j, s := range gene.Symbols {
			// fmt.Println(i, s, re4terminals.MatchString(s))
			if re4terminals.MatchString(s) {
				if v, ok := mapping[s]; ok {
					if val, okinner := indiNmsMapSliceUsed[v]; okinner {
						g.Genes[i].Symbols[j] = val[0]
					}
				} else {
					g.Genes[i].Symbols[j] = "d0"
				}
			}
		}
	}

	return indiNmSlice, nil

}

func (d *dump) generateCode() ([]byte, error) {
	var buf bytes.Buffer
	d.w = &buf
	// d.write("// GML: d.gr.Open\n")
	d.write(d.gr.Open)
	for _, h := range d.gr.Headers {
		if h.Type != "default" {
			continue
		}
		// d.write(fmt.Sprintf("// GML: d.gr.Headers: h=%#v\n", h))
		d.write(h.Chardata)
		d.write(d.gr.Endline)
	}
	for _, t := range d.gr.Tempvars {
		if t.Type != "default" {
			continue
		}
		// d.write(fmt.Sprintf("// GML: d.gr.Tempvars: t=%#v\n", t))
		d.write(t.Chardata)
		d.subs["tempvarname"] = t.Varname
		d.write(d.gr.Endline)
	}
	// Generate the expression, keeping track of any helper functions that are needed.
	helpers := make(grammars.HelperMap)
	s, ok := d.gr.Functions.FuncMap[d.genome.LinkFunc]
	if !ok {
		return nil, fmt.Errorf("unable to find grammar linking function: %v", s.Symbol())
	}
	glf, ok := s.(*grammars.Function)
	if !ok {
		return nil, fmt.Errorf("error casting link function: %v", s.Symbol())
	}
	exps := []string{""}
	for i, e := range d.genome.Genes {
		// d.write(fmt.Sprintf("// GML: d.genome.Genes: e=%#v\n", e))
		exp, err := e.Expression(d.gr, helpers)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			// d.write(fmt.Sprintf("// GML: len(d.genome.Genes)=%v\n", len(d.genome.Genes)))
			merge := strings.Replace(glf.Uniontype, "{tempvarname}", d.subs["tempvarname"], -1)
			merge = strings.Replace(merge, "{member}", exp, -1)
			merge = strings.Replace(merge, "{symbol}", glf.SymbolName, -1)
			exps = append(exps, merge)
		} else {
			// d.write(fmt.Sprintf("// GML: len(d.genome.Genes)=%v\n", len(d.genome.Genes)))
			exps = append(exps, d.subs["tempvarname"]+" = "+exp)
		}
	}
	exps = append(exps, "") // blank line
	fmt.Fprintln(d.w, strings.Join(exps, "\n"))
	for _, f := range d.gr.Footers {
		if f.Type != "default" {
			continue
		}
		// d.write(fmt.Sprintf("// GML: d.gr.Footers=%#v\n", f))
		d.write(f.Chardata)
		d.write(d.gr.Endline)
	}
	if len(helpers) > 0 { // Write out the helpers
		keys := make([]string, 0, len(helpers))
		for k := range helpers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			d.write(d.gr.Endline)
			d.write(helpers[k])
		}
	}

	clean, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), err
	}
	return clean, nil
}

func (d *dump) generateExps(indiNamesMap map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	d.w = &buf
	// Generate the expression, keeping track of any helper functions that are needed.
	helpers := make(grammars.HelperMap)
	// tmp use fmt to output the d
	fmt.Println(d)
	s, ok := d.gr.Functions.FuncMap[d.genome.LinkFunc]
	// fmt.Println(s, ok)
	if !ok {
		return nil, fmt.Errorf("unable to find grammar linking function: %v", s.Symbol())
	}
	glf, ok := s.(*grammars.Function)
	if !ok {
		return nil, fmt.Errorf("error casting link function: %v", s.Symbol())
	}
	exps := []string{""}
	for i, e := range d.genome.Genes {
		// d.write(fmt.Sprintf("// GML: d.genome.Genes: e=%#v\n", e))
		exp, err := e.Expression(d.gr, helpers)
		if err != nil {
			return nil, err
		}
		// replace the d[i] with the individual name
		for key, val := range indiNamesMap {
			exp = strings.Replace(exp, key, val, -1)
		}
		expout := []string{"Gene ", strconv.Itoa(i), " exp: ", exp}
		exps = append(exps, strings.Join(expout, ""))
	}
	linkfun := []string{"The Linkfunc: ", glf.SymbolName, "::", strconv.Itoa(glf.TerminalCount), "::", glf.Chardata}
	exps = append(exps, strings.Join(linkfun, "")) // blank line

	fmt.Fprintln(d.w, strings.Join(exps, "\n"))

	return buf.Bytes(), nil
}

func (d *dump) write(s string) {
	s = strings.Replace(s, "{CRLF}", "\n", -1)
	s = strings.Replace(s, "{TAB}", "\t", -1)
	for k, v := range d.subs {
		s = strings.Replace(s, fmt.Sprintf("{%v}", k), v, -1)
	}
	fmt.Fprint(d.w, s)
}
