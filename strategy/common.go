package strategy

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// define void
type void struct{}

type TradeSignal struct {
	Signal float64  `csv:"signal"`
	Input  []string `csv:"input"`
	Func   string   `csv:"func"`
}
// this is for the VS strategy to save the signals from csv file.
var CSVData map[string]map[string]map[string]string

func arithmeticSequence(total float64, n int, ratio float64) []float64 {
	// 1. create a slice with n elements
	slice := make([]float64, n)

	if n == 1 {
		slice[0] = total
		return slice
	} else {
		// 计算均值
		firstElement := 2.0 / (float64(n) * (1 + ratio))
		// 3. calculate the difference between the first element and the last element
		diff := firstElement*ratio - firstElement
		// 4. calculate the common difference
		d := diff / float64(n-1)
		// 5. calculate the arithmetic sequence
		for i := 0; i < n; i++ {
			slice[i] = (firstElement + float64(i)*d) * total
		}
		return slice
	}

}
// func to filter out the []SortRank with some conditions, that conditions are defined by 
// some first class function
// filterSignal is a function to filter out the signals based on certain conditions
func filterSignal(signals []SortRank, filterSignal func(SortRank) bool) []SortRank {
	var res []SortRank
	for _, signal := range signals {
		if filterSignal(signal) {
			res = append(res, signal)
		}
	}
	return res
}
func filterSignalVS(signals []SortRankVS, filterSignal func(SortRankVS) bool) []SortRankVS {
	var res []SortRankVS
	for _, signal := range signals {
		if filterSignal(signal) {
			res = append(res, signal)
		}
	}
	return res
}



// the two functions below are used to filter out the signals based on the conditions
// normally, we will do it in the fomular section
// but here we define in the strategy section
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

// // Converts the value from a string to float64
func TryConvertToFloat(val string) (float64, error) {
	value, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// ContainNaN 此处是为了停盘数据处理设定的规则相检查用的
func ContainNaN(m dataframe.StreamingRecord) bool {
	for _, x := range m.Data {
		if len(x) == 0 {
			return true
		}
	}
	return false
}

