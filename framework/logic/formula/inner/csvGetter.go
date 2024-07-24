package indicator

import (
	"encoding/csv"
	"os"
	"strings"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// map[id]map[date]map[indi]value
var CSVData map[string]map[string]map[string]string



// !!!!!! ugly code here
var srce string = "/Users/wonderstone/Desktop/QuantKit/EX-VS-dir/download/1dayfactor/to_gep.csv" 
var InstColName string ="data_code"
var DateColName ="end_date"


// CR is the csv reader indicator
type CSVGetter struct {
	Name        string
	// Source      string
	DateColName string
	InstColName string
	IndiName    string
	InstID      string
	lastval     string
	initState   bool
	Mode 		string
}


func (g *CSVGetter) DoInit(f config.Formula) {
	g.Name = f.Name
	g.InstID = f.InstID
	// g.Source = config.MustGetParamString(f.Param, "Source")
	g.DateColName = config.MustGetParamString(f.Param, "DateColName")
	g.InstColName = config.MustGetParamString(f.Param, "InstColName")
	g.IndiName = config.MustGetParamString(f.Param, "IndiName")
	g.Mode = config.MustGetParamString(f.Param, "Mode")
	g.lastval = ""
}

func (g *CSVGetter) DoCalculate(tm time.Time, row dataframe.RecordFunc) string {
	// try to get the value from map
	// turn tm to string with format "2006.01.02"
	tmStr := tm.Format("2006.01.02")
	if g.Mode == "LV"{
		// this is the last value mode
		// compare the value with the last value
		// if the value is equal to the last value return the ""
		// if the value is not equal to the last value return the value
		if v, ok := CSVData[tmStr]; ok {
			if v1, ok1 := v[ModifyInstID(g.InstID)]; ok1 {
				if v2, ok2 := v1[g.IndiName]; ok2 {
					if g.lastval == v2 {
						return ""
					}
					g.lastval = v2
					return v2
				}
			}
			// if the value is not in the map return the last value
			} else {
				return ""
			}
		// if the value is not in the map return the last value
		return ""

	}else{
		// this is the normal mode
		if v, ok := CSVData[tmStr]; ok {
			if v1, ok1 := v[ModifyInstID(g.InstID)]; ok1 {
				if v2, ok2 := v1[g.IndiName]; ok2 {
					g.lastval = v2
					return v2
				}
			}
			// if the value is not in the map return the last value
			} else {
				return ""
			}
			return ""
	}
}

func (m *CSVGetter) DoReset() {

}

// NewMA returns a new MA indicator
func NewCSVGetter(Name string,	DateColName string,	InstColName string,
	IndiName string,InstID string,Mode string) *CSVGetter {
	return &CSVGetter{
		Name:   Name,
		// Source: Source,
		DateColName: DateColName,
		InstColName: InstColName,
		IndiName: IndiName,
		InstID: InstID,
		Mode: Mode,
	}
}

// func to read csv file and return a map[string]string
func ProcCsvBase(source string, InstColName, DateColName string)  map[string]map[string]map[string]string {
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

	// create a map to store the data map[string]map[string]map[string]string
	dataMap := make(map[string]map[string]map[string]string)



	// read all the records
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// only the record that InstColName is equal to InstID add the IndiName data to the map

	for _, recd := range records {
		// check the value of dataMap[record[dateCC]] has key record[instCC] or not
		// if not create a new map[string]string
		if dataMap[recd[dateCC]] == nil {
			dataMap[recd[dateCC]] = make(map[string]map[string]string)
		}
		if dataMap[recd[dateCC]][recd[instCC]] == nil {
			dataMap[recd[dateCC]][recd[instCC]] = make(map[string]string)
		}
		// add the data to the map
		for i, v := range recd {
			if i != dateCC && i != instCC {
			dataMap[recd[dateCC]][recd[instCC]][record[i]] = v
			}
		}
	}
	
	return dataMap
}


func init() {
	formula.RegisterNewFormula(new(CSVGetter), "CSVGetter")
	CSVData = ProcCsvBase(srce, InstColName, DateColName)

}


// func to modify the instrID from 000019.XSHE.CS or 600105.XSHG.CS to 000019.SZ or 600105.SH
func ModifyInstID(instID string) string {
	// use . to split the instID
	inst := strings.Split(instID, ".")
	if inst[1] == "XSHE" {
		return inst[0] + ".SZ"
		} else {
			return inst[0] + ".SH"
		}
}