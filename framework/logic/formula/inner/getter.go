// All rights reserved. This is part of West Securities ltd. proprietary source code.
// No part of this file may be reproduced or transmitted in any form or by any means,
// electronic or mechanical, including photocopying, recording, or by any information
// storage and retrieval system, without the prior written permission of West Securities ltd.

// author:  Wonderstone (Digital Office Product Department #2)
// revisor:

package indicator

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// MA is the moving average indicator
type Getter struct {
	Name        string
	Source      string
	DateColName string
	InstColName string
	IndiName    string
	InstID      string
	lastval     string
	Data        map[string]string
}

func (g *Getter) DoInit(f config.Formula) {
	g.Name = f.Name
	g.InstID = f.InstID
	g.Source = config.MustGetParamString(f.Param, "Source")
	g.DateColName = config.MustGetParamString(f.Param, "DateColName")
	g.InstColName = config.MustGetParamString(f.Param, "InstColName")
	g.IndiName = config.MustGetParamString(f.Param, "IndiName")
	g.lastval = ""

	g.Data = ProcCsv(g.Source, g.InstColName, g.DateColName, g.InstID, g.IndiName)

}

func (g *Getter) DoCalculate(tm time.Time, row dataframe.RecordFunc) string {
	// try to get the value from map
	// turn tm to string with format "2006.01.02"
	tmStr := tm.Format("2006.01.02")

	if v, ok := g.Data[tmStr]; ok {
		out := fmt.Sprintf("%.4s", v)
		g.lastval = out
		return out
	} else {
		return g.lastval
	}

}

func (m *Getter) DoReset() {

}

// NewMA returns a new MA indicator
func NewGetter(Name string, Source string,	DateColName string,	InstColName string,
	IndiName string,InstID string) *Getter {
	return &Getter{
		Name:   Name,
		InstID: InstID,
		Source: Source,
		DateColName: DateColName,
		InstColName: InstColName,
		IndiName: IndiName,
		Data:   ProcCsv(Source, InstColName, DateColName, InstID, IndiName),
	}
}

func init() {
	formula.RegisterNewFormula(new(Getter), "Getter")
}

// func to read csv file and return a map[string]string
func ProcCsv(source string, InstColName, DateColName string, InstID, IndiName string) map[string]string {
	file, err := os.Open(source)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// create a new reader
	reader := csv.NewReader(file)
	// read the first record and save it to the variable
	record, err := reader.Read()
	if err != nil {
		panic(err)
	}

	// check the location of the InstColName, if it is not in the record panic
	instCC := CheckLocation(record, InstColName)
	// check the location of the DateCol, if it is not in the record panic
	dateCC := CheckLocation(record, DateColName)
	// check the location of the IndiName, if it is not in the record panic
	indiCC := CheckLocation(record, IndiName)

	// create a map to store the data
	dataMap := make(map[string]string)

	// read all the records
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// only the record that InstColName is equal to InstID add the IndiName data to the map

	for _, record := range records {
		// only the first 6 characters is used for now
		if record[instCC][0:6] ==  InstID[:6] {
			dataMap[record[dateCC]] = record[indiCC]
		}
	}

	return dataMap
}

func CheckLocation(record []string, colName string) int {
	colIndex := -1
	for i, col := range record {
		if col == colName {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		panic("InstColName not found in the record")
	}
	return colIndex
}

