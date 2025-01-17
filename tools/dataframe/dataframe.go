package dataframe

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"golang.org/x/exp/slices"
)

type Record struct {
	Data []string
}

type DataFrame struct {
	FrameRecords  []Record
	HeaderToIndex map[string]int
	Header        []string
}

type StreamingRecord struct {
	Data    []string
	Headers map[string]int
}

type RecordFunc interface {
	Val(fieldName string, header ...map[string]int) string
	Update(fieldName, value string, header ...map[string]int)
}

func ConvertToFloat(recordFunc RecordFunc, fieldName string, header ...map[string]int) float64 {
	value, err := strconv.ParseFloat(recordFunc.Val(fieldName, header...), 64)
	if err != nil {
		config.ErrorF("不能转换为 float64: %v", err)
	}
	return value
}

func TryConvertToFloat(recordFunc RecordFunc, fieldName string, header ...map[string]int) (float64, error) {
	value, err := strconv.ParseFloat(recordFunc.Val(fieldName, header...), 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ConvertToInt(recordFunc RecordFunc, fieldName string, header ...map[string]int) int64 {
	value, err := strconv.ParseInt(recordFunc.Val(fieldName, header...), 0, 64)
	if err != nil {
		config.ErrorF("不能转换为 int64: %v", err)
	}
	return value
}

func TryConvertToInt(recordFunc RecordFunc, fieldName string, header ...map[string]int) (int64, error) {
	value, err := strconv.ParseInt(recordFunc.Val(fieldName, header...), 0, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (x StreamingRecord) Update(fieldName, value string, header ...map[string]int) {
	if _, ok := x.Headers[fieldName]; !ok {
		config.ErrorF("提供的字段 %s 不是数据帧中的有效字段。", fieldName)
	}

	x.Data[x.Headers[fieldName]] = value
}

// Return the value of the specified field.
func (x StreamingRecord) Val(fieldName string, header ...map[string]int) string {
	if _, ok := x.Headers[fieldName]; !ok {
		panic(fmt.Errorf("提供的指标 %s 不存在", fieldName))
	}
	return x.Data[x.Headers[fieldName]]
}

// Converts the value from a string to float64
func (x StreamingRecord) ConvertToFloat(fieldName string) float64 {
	value, err := strconv.ParseFloat(x.Val(fieldName), 64)
	if err != nil {
		config.ErrorF("不能转换为 float64: %v", err)
	}
	return value
}

// Converts the value from a string to float64
func (x StreamingRecord) TryConvertToFloat(fieldName string) (float64, error) {
	value, err := strconv.ParseFloat(x.Val(fieldName), 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// Converts the value from a string to int64
func (x StreamingRecord) ConvertToInt(fieldName string) int64 {
	value, err := strconv.ParseInt(x.Val(fieldName), 0, 64)
	if err != nil {
		config.ErrorF("不能转换为 int64: %v", err)
	}
	return value
}

// Generate a new empty DataFrame.
func CreateNewDataFrame(headers []string) DataFrame {
	var myRecords []Record
	theHeaders := make(map[string]int)

	// Add headers to map in correct order
	for i := 0; i < len(headers); i++ {
		theHeaders[headers[i]] = i
	}

	newFrame := DataFrame{FrameRecords: myRecords, HeaderToIndex: theHeaders, Header: headers}

	return newFrame
}

// Generate a new DataFrame sourced from a csv file.
func CreateDataFrame(dir, fileName string, column ...string) DataFrame {
	columns := map[string]any{}

	for _, col := range column {
		columns[col] = nil
	}

	if !strings.HasSuffix(fileName, ".csv") {
		fileName = fileName + ".csv"
	}

	// Open the CSV file
	recordFile, err := os.Open(path.Join(dir, fileName))
	if err != nil {
		config.ErrorF("打开文件失败: %s，请确认文件路径无误", path.Join(dir, fileName))
	}

	// Setup the reader
	reader := csv.NewReader(recordFile)

	// Read the records
	header, err := reader.Read()
	if err != nil {
		config.WarnF("读取文件 %s 记录失败: %v", path.Join(dir, fileName), err)
		return CreateNewDataFrame([]string{})
	}

	// Remove Byte Order Marker for UTF-8 files
	for i, each := range header {
		byteSlice := []byte(each)
		if byteSlice[0] == 239 && byteSlice[1] == 187 && byteSlice[2] == 191 {
			// 检查是否为正确的列
			if _, ok := columns[each[3:]]; !ok {
				config.ErrorF("字段 %s 不在 %v 之中", each[3:], columns)
			}
			header[i] = each[3:]
		}
	}

	headers := make(map[string]int)
	for i, columnName := range header {
		headers[columnName] = i
	}

	// Empty slice to store Records
	var s []Record

	// Loop over the records and create Record objects to be stored
	for i := 0; ; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			config.ErrorF("读取文件 %s 出错: %v", path.Join(dir, fileName), err)
		}
		// Create new Record
		x := Record{Data: []string{}}

		// Loop over records and add to Data field of Record struct
		for _, r := range record {
			x.Data = append(x.Data, r)
		}
		s = append(s, x)
	}
	newFrame := DataFrame{FrameRecords: s, HeaderToIndex: headers, Header: header}
	return newFrame
}

// Stream rows of data from a csv file to be processed. Streaming data is preferred when dealing with large files
// and memory usage needs to be considered. Results are streamed via a channel with a StreamingRecord type.
func Stream(dir, fileName string, c chan StreamingRecord) {
	defer close(c)

	if !strings.HasSuffix(fileName, ".csv") {
		fileName = fileName + ".csv"
	}

	// Open the CSV file
	recordFile, err := os.Open(path.Join(dir, fileName))
	if err != nil {
		config.ErrorF("打开文件失败: %s，请确认文件路径无误", path.Join(dir, fileName))
	}

	// Setup the reader
	reader := csv.NewReader(recordFile)

	// Read the records
	header, err := reader.Read()
	if err != nil {
		config.ErrorF("读取文件失败: %s，请确认文件格式无误", path.Join(dir, fileName))
	}

	// Remove Byte Order Marker for UTF-8 files
	for i, each := range header {
		byteSlice := []byte(each)
		if byteSlice[0] == 239 && byteSlice[1] == 187 && byteSlice[2] == 191 {
			header[i] = each[3:]
		}
	}

	headers := make(map[string]int)
	for i, columnName := range header {
		headers[columnName] = i
	}

	// Loop over the records and create Record objects to be stored
	for i := 0; ; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			config.ErrorF("读取文件失败: %v", err)
		}
		// Create new Record
		x := StreamingRecord{Headers: headers}

		// Loop over records and add to Data field of Record struct
		for _, r := range record {
			x.Data = append(x.Data, r)
		}
		c <- x
	}

	return
}

func WriteStream(
	dir, fileName string, c chan StreamingRecord, header []string, handler func(record StreamingRecord) []string,
) {
	if !strings.HasSuffix(fileName, ".csv") {
		fileName = fileName + ".csv"
	}

	// Create the csv file
	recordFile, err := os.Create(path.Join(dir, fileName))
	defer recordFile.Close()
	if err != nil {
		config.ErrorF("创建csv文件失败: %v", err)
	}

	// 初始化writer
	writer := csv.NewWriter(recordFile)

	// 写入header
	err = writer.Write(header)
	if err != nil {
		config.ErrorF("写入记录失败: %v", err)
		return
	}

	// 写入数据
	for x := range c {
		err := writer.Write(handler(x))
		if err != nil {
			config.ErrorF("写入记录失败: %v", err)
			return
		}
	}

	writer.Flush()
}

func worker(
	jobs <-chan string, results chan<- DataFrame, resultsNames chan<- string, filePath string, column ...string,
) {
	for n := range jobs {
		df := CreateDataFrame(filePath, n, column...)
		results <- df
		resultsNames <- n
	}
}

// Concurrently loads multiple csv files into DataFrames within the same directory.
// Returns a slice with the DataFrames in the same order as provided in the files parameter.
func LoadFrames(filePath string, files []string, column ...string) ([]DataFrame, error) {
	numJobs := len(files)

	// if numJobs <= 1 {
	// 	return nil, errors.New("LoadFrames requires at least two files")
	// }

	jobs := make(chan string, numJobs)
	results := make(chan DataFrame, numJobs)
	resultsNames := make(chan string, numJobs)

	// Generate workers
	for i := 0; i < 4; i++ {
		go worker(jobs, results, resultsNames, filePath, column...)
	}

	// Load up the jobs channel
	for i := 0; i < numJobs; i++ {
		jobs <- files[i]
	}
	close(jobs) // Close jobs channel once loaded

	// Map to store results
	jobResults := make(map[string]DataFrame)

	// Collect results and store in map
	for i := 1; i <= numJobs; i++ {
		jobResults[<-resultsNames] = <-results
	}

	var orderedResults []DataFrame
	for _, f := range files {
		val, ok := jobResults[f]
		if !ok {
			return []DataFrame{}, errors.New("an error occurred while looking up returned DataFrame in the LoadFrames function")
		}
		orderedResults = append(orderedResults, val)
	}
	return orderedResults, nil
}

// KeepColumns User specifies columns they want to keep from a preexisting DataFrame
func (frame DataFrame) KeepColumns(columns []string) DataFrame {
	df := CreateNewDataFrame(columns)

	for _, row := range frame.FrameRecords {
		var newData []string
		for _, column := range columns {
			newData = append(newData, row.Val(column, frame.HeaderToIndex))
		}
		df = df.AddRecord(newData)
	}

	return df
}

// User specifies columns they want to remove from a preexisting DataFrame
func (frame DataFrame) RemoveColumns(columns ...string) DataFrame {
	approvedColumns := []string{}

	for _, col := range frame.Columns() {
		if !slices.Contains(columns, col) {
			approvedColumns = append(approvedColumns, col)
		}
	}

	return frame.KeepColumns(approvedColumns)
}

// Rename a specified column in the DataFrame
func (frame *DataFrame) Rename(originalColumnName, newColumnName string) error {
	columns := []string{}
	var columnLocation int

	for k, v := range frame.HeaderToIndex {
		columns = append(columns, k)
		if k == originalColumnName {
			columnLocation = v
		}
	}

	// Check original column name is found in DataFrame
	if !slices.Contains(columns, originalColumnName) {
		return errors.New("The original column name provided was not found in the DataFrame")
	}

	// Check new column name does not already exist
	if slices.Contains(columns, newColumnName) {
		return errors.New("The provided new column name already exists in the DataFrame and is not allowed")
	}

	// Remove original column name
	delete(frame.HeaderToIndex, originalColumnName)

	// Add new column name
	frame.HeaderToIndex[newColumnName] = columnLocation

	return nil
}

// Add a new record to the DataFrame
func (frame DataFrame) AddRecord(newData []string) DataFrame {
	x := Record{Data: []string{}}

	for _, each := range newData {
		x.Data = append(x.Data, each)
	}

	frame.FrameRecords = append(frame.FrameRecords, x)

	return frame
}

// Provides a slice of columns in order
func (frame DataFrame) Columns() []string {
	var columns []string

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				columns = append(columns, k)
			}
		}
	}
	return columns
}

// Generates a decoupled copy of an existing DataFrame.
// Changes made to either the original or new copied frame
// will not be reflected in the other.
func (frame DataFrame) Copy() DataFrame {
	headers := []string{}

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				headers = append(headers, k)
			}
		}
	}
	df := CreateNewDataFrame(headers)

	for i := 0; i < len(frame.FrameRecords); i++ {
		df = df.AddRecord(frame.FrameRecords[i].Data)
	}
	return df
}

// Generates a new filtered DataFrame.
// New DataFrame will be kept in same order as original.
func (frame DataFrame) Filtered(fieldName string, value ...string) DataFrame {
	headers := []string{}

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				headers = append(headers, k)
			}
		}
	}
	newFrame := CreateNewDataFrame(headers)

	for i := 0; i < len(frame.FrameRecords); i++ {
		if slices.Contains(value, frame.FrameRecords[i].Data[frame.HeaderToIndex[fieldName]]) {
			newFrame = newFrame.AddRecord(frame.FrameRecords[i].Data)
		}
	}

	return newFrame
}

// Generated a new filtered DataFrame that in which a numerical column is either greater than or equal to
// a provided numerical value.
func (frame DataFrame) GreaterThanOrEqualTo(fieldName string, value float64) (DataFrame, error) {
	headers := []string{}

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				headers = append(headers, k)
			}
		}
	}
	newFrame := CreateNewDataFrame(headers)

	for i, row := range frame.FrameRecords {
		valString := row.Val(fieldName, frame.HeaderToIndex)

		val, err := strconv.ParseFloat(valString, 64)
		if err != nil {
			return CreateNewDataFrame([]string{}), err
		}

		if val >= value {
			newFrame = newFrame.AddRecord(frame.FrameRecords[i].Data)
		}
	}
	return newFrame, nil
}

// Generated a new filtered DataFrame that in which a numerical column is either less than or equal to
// a provided numerical value.
func (frame DataFrame) LessThanOrEqualTo(fieldName string, value float64) (DataFrame, error) {
	headers := []string{}

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				headers = append(headers, k)
			}
		}
	}
	newFrame := CreateNewDataFrame(headers)

	for i, row := range frame.FrameRecords {
		valString := row.Val(fieldName, frame.HeaderToIndex)

		val, err := strconv.ParseFloat(valString, 64)
		if err != nil {
			return CreateNewDataFrame([]string{}), err
		}

		if val <= value {
			newFrame = newFrame.AddRecord(frame.FrameRecords[i].Data)
		}
	}
	return newFrame, nil
}

// Generates a new DataFrame that excludes specified instances.
// New DataFrame will be kept in same order as original.
func (frame DataFrame) Exclude(fieldName string, value ...string) DataFrame {
	headers := []string{}

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				headers = append(headers, k)
			}
		}
	}
	newFrame := CreateNewDataFrame(headers)

	for i := 0; i < len(frame.FrameRecords); i++ {
		if !slices.Contains(value, frame.FrameRecords[i].Data[frame.HeaderToIndex[fieldName]]) {
			newFrame = newFrame.AddRecord(frame.FrameRecords[i].Data)
		}
	}

	return newFrame
}

// Generates a new filtered DataFrame with all records occuring after a specified date provided by the user.
// User must provide the date field as well as the desired date.
// Instances where record dates occur on the same date provided by the user will not be included.
// Records must occur after the specified date.
func (frame DataFrame) FilteredAfter(fieldName, desiredDate string) DataFrame {
	headers := []string{}

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				headers = append(headers, k)
			}
		}
	}
	newFrame := CreateNewDataFrame(headers)

	for i := 0; i < len(frame.FrameRecords); i++ {
		recordDate := dateConverter(frame.FrameRecords[i].Data[frame.HeaderToIndex[fieldName]])
		isAfter := recordDate.After(dateConverter(desiredDate))

		if isAfter {
			newFrame = newFrame.AddRecord(frame.FrameRecords[i].Data)
		}
	}
	return newFrame
}

// Generates a new filtered DataFrame with all records occuring before a specified date provided by the user.
// User must provide the date field as well as the desired date.
// Instances where record dates occur on the same date provided by the user will not be included. Records must occur
// before the specified date.
func (frame DataFrame) FilteredBefore(fieldName, desiredDate string) DataFrame {
	headers := []string{}

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				headers = append(headers, k)
			}
		}
	}
	newFrame := CreateNewDataFrame(headers)

	for i := 0; i < len(frame.FrameRecords); i++ {
		recordDate := dateConverter(frame.FrameRecords[i].Data[frame.HeaderToIndex[fieldName]])
		isBefore := recordDate.Before(dateConverter(desiredDate))

		if isBefore {
			newFrame = newFrame.AddRecord(frame.FrameRecords[i].Data)
		}
	}

	return newFrame
}

// Generates a new filtered DataFrame with all records occuring between a specified date range provided by the user.
// User must provide the date field as well as the desired date.
// Instances where record dates occur on the same date provided by the user will not be included. Records must occur
// between the specified start and end dates.
func (frame DataFrame) FilteredBetween(fieldName, startDate, endDate string) DataFrame {
	headers := []string{}

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				headers = append(headers, k)
			}
		}
	}
	newFrame := CreateNewDataFrame(headers)

	for i := 0; i < len(frame.FrameRecords); i++ {
		recordDate := dateConverter(frame.FrameRecords[i].Data[frame.HeaderToIndex[fieldName]])
		isAfter := recordDate.After(dateConverter(startDate))
		isBefore := recordDate.Before(dateConverter(endDate))

		if isAfter && isBefore {
			newFrame = newFrame.AddRecord(frame.FrameRecords[i].Data)
		}
	}

	return newFrame
}

// Creates a new field and assigns and empty string.
func (frame *DataFrame) NewField(fieldName string) {
	for i, _ := range frame.FrameRecords {
		frame.FrameRecords[i].Data = append(frame.FrameRecords[i].Data, "")
	}
	frame.HeaderToIndex[fieldName] = len(frame.HeaderToIndex)
}

// Return a slice of all unique values found in a specified field.
func (frame *DataFrame) Unique(fieldName string) []string {
	var results []string

	for _, row := range frame.FrameRecords {
		if !slices.Contains(results, row.Val(fieldName, frame.HeaderToIndex)) {
			results = append(results, row.Val(fieldName, frame.HeaderToIndex))
		}
	}
	return results
}

// Stack two DataFrames with matching headers.
func (frame DataFrame) ConcatFrames(dfNew *DataFrame) (DataFrame, error) {
	// Check number of columns in each frame match.
	if len(frame.HeaderToIndex) != len(dfNew.HeaderToIndex) {
		return frame, errors.New("Cannot ConcatFrames as columns do not match.")
	}

	// Check columns in both frames are in the same order.
	originalFrame := []string{}
	for i := 0; i <= len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				originalFrame = append(originalFrame, k)
			}
		}
	}

	newFrame := []string{}
	for i := 0; i <= len(dfNew.HeaderToIndex); i++ {
		for k, v := range dfNew.HeaderToIndex {
			if v == i {
				newFrame = append(newFrame, k)
			}
		}
	}

	for i, each := range originalFrame {
		if each != newFrame[i] {
			return frame, errors.New("Cannot ConcatFrames as columns are not in the same order.")
		}
	}

	// Iterate over new dataframe in order
	for i := 0; i < len(dfNew.FrameRecords); i++ {
		frame.FrameRecords = append(frame.FrameRecords, dfNew.FrameRecords[i])
	}
	return frame, nil
}

// Import all columns from right frame into left frame if no columns
// are provided by the user. Process must be done so in order.
func (frame DataFrame) Merge(dfRight *DataFrame, primaryKey string, columns ...string) {
	if len(columns) == 0 {
		for i := 0; i < len(dfRight.HeaderToIndex); i++ {
			for k, v := range dfRight.HeaderToIndex {
				if v == i {
					columns = append(columns, k)
				}
			}
		}
	} else {
		// Ensure columns user provided are all found in right frame.
		for _, col := range columns {
			colStatus := false
			for k, _ := range dfRight.HeaderToIndex {
				if col == k {
					colStatus = true
				}
			}
			// Ensure there are no duplicated columns other than the primary key.
			if colStatus != true {
				panic("Merge Error: User provided column not found in right dataframe.")
			}
		}
	}

	// Check that no columns are duplicated between the two frames (other than primaryKey).
	for _, col := range columns {
		for k, _ := range frame.HeaderToIndex {
			if col == k && col != primaryKey {
				panic("The following column is duplicated in both frames and is not the specified primary key which is not allowed: " + col)
			}
		}
	}

	// Load map indicating the location of each lookup value in right frame.
	lookup := make(map[string]int)
	for i, row := range dfRight.FrameRecords {
		lookup[row.Val(primaryKey, dfRight.HeaderToIndex)] = i
	}

	// Create new columns in left frame.
	for _, col := range columns {
		if col != primaryKey {
			frame.NewField(col)
		}
	}

	// Iterate over left frame and add new data.
	for _, row := range frame.FrameRecords {
		lookupVal := row.Val(primaryKey, frame.HeaderToIndex)

		if val, ok := lookup[lookupVal]; ok {
			for _, col := range columns {
				if col != primaryKey {
					valToAdd := dfRight.FrameRecords[val].Data[dfRight.HeaderToIndex[col]]
					row.Update(col, valToAdd, frame.HeaderToIndex)
				}
			}
		}
	}
}

// Performs an inner merge where all columns are consolidated between the two frames but only for records
// where the specified primary key is found in both frames.
func (frame DataFrame) InnerMerge(dfRight *DataFrame, primaryKey string) DataFrame {
	var rightFrameColumns []string

	for i := 0; i < len(dfRight.HeaderToIndex); i++ {
		for k, v := range dfRight.HeaderToIndex {
			if v == i {
				rightFrameColumns = append(rightFrameColumns, k)
			}
		}
	}

	var leftFrameColumns []string

	for i := 0; i < len(frame.HeaderToIndex); i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				leftFrameColumns = append(leftFrameColumns, k)
			}
		}
	}

	// Ensure the specified primary key is found in both frames.
	var lStatus bool
	var rStatus bool

	for _, col := range leftFrameColumns {
		if col == primaryKey {
			lStatus = true
		}
	}

	for _, col := range rightFrameColumns {
		if col == primaryKey {
			rStatus = true
		}
	}

	if !lStatus || !rStatus {
		panic("The specified primary key was not found in both DataFrames.")
	}

	// Find position of primary key column in right frame.
	var rightFramePrimaryKeyPosition int
	for i, col := range rightFrameColumns {
		if col == primaryKey {
			rightFramePrimaryKeyPosition = i
		}
	}

	// Check that no columns are duplicated between the two frames (other than primaryKey).
	for _, col := range rightFrameColumns {
		for k, _ := range frame.HeaderToIndex {
			if col == k && col != primaryKey {
				panic("The following column is duplicated in both frames and is not the specified primary key which is not allowed: " + col)
			}
		}
	}

	// Load map indicating the location of each lookup value in right frame.
	rLookup := make(map[string]int)
	for i, row := range dfRight.FrameRecords {
		// Only add if key hasn't already been added. This ensures the first record found in the right
		// frame is what is used instead of the last if duplicates are found.
		currentKey := row.Val(primaryKey, dfRight.HeaderToIndex)
		_, ok := rLookup[currentKey]
		if !ok {
			rLookup[currentKey] = i
		}
	}

	// New DataFrame to house records found in both frames.
	dfNew := CreateNewDataFrame(leftFrameColumns)

	// Add right frame columns to new DataFrame.
	for i, col := range rightFrameColumns {
		// Skip over primary key column in right frame as it was already included in the left frame.
		if i != rightFramePrimaryKeyPosition {
			dfNew.NewField(col)
		}
	}

	var approvedPrimaryKeys []string

	// Create slice of specified RID's found in both frames.
	for _, lRow := range frame.FrameRecords {
		currentKey := lRow.Val(primaryKey, frame.HeaderToIndex)

		// Skip blank values as they are not allowed.
		if len(currentKey) == 0 || strings.ToLower(currentKey) == "nan" || strings.ToLower(currentKey) == "null" {
			continue
		}

		for _, rRow := range dfRight.FrameRecords {
			currentRightFrameKey := rRow.Val(primaryKey, dfRight.HeaderToIndex)
			// Add primary key to approved list if found in right frame.
			if currentRightFrameKey == currentKey {
				approvedPrimaryKeys = append(approvedPrimaryKeys, currentKey)
			}
		}
	}

	// Add approved records to new DataFrame.
	for i, row := range frame.FrameRecords {
		currentKey := row.Val(primaryKey, frame.HeaderToIndex)
		if slices.Contains(approvedPrimaryKeys, currentKey) {
			lData := frame.FrameRecords[i].Data
			rData := dfRight.FrameRecords[rLookup[currentKey]].Data

			// Add left frame data to variable.
			var data []string
			data = append(data, lData...)

			// Add all right frame data while skipping over the primary key column.
			// The primary key column is skipped as it has already been added from the left frame.
			for i, d := range rData {
				if i != rightFramePrimaryKeyPosition {
					data = append(data, d)
				}
			}

			dfNew = dfNew.AddRecord(data)
		}
	}
	return dfNew
}

func (frame *DataFrame) CountRecords() int {
	return len(frame.FrameRecords)
}

// Return a sum of float64 type of a numerical field.
func (frame *DataFrame) Sum(fieldName string) float64 {
	var sum float64

	for _, row := range frame.FrameRecords {
		val, err := strconv.ParseFloat(row.Val(fieldName, frame.HeaderToIndex), 64)
		if err != nil {
			config.ErrorF("Could Not Convert String to Float During Sum: %v", err)
		}
		sum += val
	}
	return sum
}

// Return an average of type float64 of a numerical field.
func (frame *DataFrame) Average(fieldName string) float64 {
	sum := frame.Sum(fieldName)
	count := frame.CountRecords()

	if count == 0 {
		return 0.0
	}
	return sum / float64(count)
}

// Return the maximum value in a numerical field.
func (frame *DataFrame) Max(fieldName string) float64 {
	maximum := 0.0
	for i, row := range frame.FrameRecords {
		// Set the max to the first value in dataframe.
		if i == 0 {
			initialMax, err := strconv.ParseFloat(row.Val(fieldName, frame.HeaderToIndex), 64)
			if err != nil {
				config.ErrorF("Could Not Convert String to Float During Sum: %v", err)
			}
			maximum = initialMax
		}
		val, err := strconv.ParseFloat(row.Val(fieldName, frame.HeaderToIndex), 64)
		if err != nil {
			config.ErrorF("Could Not Convert String to Float During Sum: %v", err)
		}

		if val > maximum {
			maximum = val
		}
	}
	return maximum
}

// Return the minimum value in a numerical field.
func (frame *DataFrame) Min(fieldName string) float64 {
	min := 0.0
	for i, row := range frame.FrameRecords {
		// Set the max to the first value in dataframe.
		if i == 0 {
			initialMin, err := strconv.ParseFloat(row.Val(fieldName, frame.HeaderToIndex), 64)
			if err != nil {
				config.ErrorF("Could Not Convert String to Float During Sum: %v", err)
			}
			min = initialMin
		}
		val, err := strconv.ParseFloat(row.Val(fieldName, frame.HeaderToIndex), 64)
		if err != nil {
			config.ErrorF("Could Not Convert String to Float During Sum: %v", err)
		}

		if val < min {
			min = val
		}
	}
	return min
}

func standardDeviation(num []float64) float64 {
	l := float64(len(num))
	sum := 0.0
	var sd float64

	for _, n := range num {
		sum += n
	}

	mean := sum / l

	for j := 0; j < int(l); j++ {
		// The use of Pow math function func Pow(x, y float64) float64
		sd += math.Pow(num[j]-mean, 2)
	}
	// The use of Sqrt math function func Sqrt(x float64) float64
	sd = math.Sqrt(sd / l)

	return sd
}

// Return the standard deviation of a numerical field.
func (frame *DataFrame) StandardDeviation(fieldName string) (float64, error) {
	var nums []float64

	for _, row := range frame.FrameRecords {
		num, err := strconv.ParseFloat(row.Val(fieldName, frame.HeaderToIndex), 64)
		if err != nil {
			return 0.0, errors.New("Could not convert string to number in specified column to calculate standard deviation.")
		}
		nums = append(nums, num)
	}
	return standardDeviation(nums), nil
}

func (frame *DataFrame) SaveDataFrame(dir, fileName string) bool {
	if !strings.HasSuffix(fileName, ".csv") {
		fileName = fileName + ".csv"
	}

	// Create the csv file
	csvFile, err := os.Create(path.Join(dir, fileName))
	defer csvFile.Close()
	if err != nil {
		config.ErrorF("Error creating the blank csv file to save the data: %v", err)
	}

	w := csv.NewWriter(csvFile)
	defer w.Flush()

	var data [][]string
	var row []string
	columnLength := len(frame.HeaderToIndex)

	// Write headers to top of file
	for i := 0; i < columnLength; i++ {
		for k, v := range frame.HeaderToIndex {
			if v == i {
				row = append(row, k)
			}
		}
	}
	data = append(data, row)

	// Add Data
	for i := 0; i < len(frame.FrameRecords); i++ {
		var row []string
		for pos := 0; pos < columnLength; pos++ {
			row = append(row, frame.FrameRecords[i].Data[pos])
		}
		data = append(data, row)
	}

	w.WriteAll(data)

	return true
}

// Val Return the value of the specified field.
func (x Record) Val(fieldName string, headers ...map[string]int) string {
	if len(headers) == 0 {
		config.ErrorF("需要提供标头信息")
	}

	if header, ok := headers[0][fieldName]; !ok {
		config.ErrorF("提供的字段 %s 不是数据帧中的有效字段。", fieldName)
} else {
		return x.Data[header]
	}
	
	return ""
}

// Update the value in a specified field.
func (x Record) Update(fieldName, value string, headers ...map[string]int) {
	if len(headers) == 0 {
		config.ErrorF("需要提供标头信息")
	}

	if header, ok := headers[0][fieldName]; !ok {
		config.ErrorF("提供的字段 %s 不是数据帧中的有效字段。", fieldName)
} else {
		x.Data[header] = value
	}
	}

// ConvertToFloat Converts the value from a string to float64.
func (x Record) ConvertToFloat(fieldName string, headers map[string]int) float64 {
	value, err := strconv.ParseFloat(x.Val(fieldName, headers), 64)
	if err != nil {
		config.ErrorF("不能转换为 float64: %v", err)
	}
	return value
}

// Converts the value from a string to int64.
func (x Record) ConvertToInt(fieldName string, headers map[string]int) int64 {
	value, err := strconv.ParseInt(x.Val(fieldName, headers), 0, 64)
	if err != nil {
		config.ErrorF("不能转换为 int64: %v", err)
	}
	return value
}

// Converts various date strings into time.Time
func dateConverter(dateString string) time.Time {
	// Convert date if not in 2006-01-02 format
	if strings.Contains(dateString, "/") {
		dateSlice := strings.Split(dateString, "/")

		if len(dateSlice[0]) != 2 {
			dateSlice[0] = "0" + dateSlice[0]
		}
		if len(dateSlice[1]) != 2 {
			dateSlice[1] = "0" + dateSlice[1]
		}
		if len(dateSlice[2]) == 2 {
			dateSlice[2] = "20" + dateSlice[2]
		}
		dateString = dateSlice[2] + "-" + dateSlice[0] + "-" + dateSlice[1]
	}

	value, err := time.ParseInLocation("2006-01-02", dateString, time.Local)
	if err != nil {
		config.ErrorF("不能转换为 time.Time: %v", err)
	}
	return value
}

// Converts date from specified field to time.Time
func (x Record) ConvertToDate(fieldName string, headers map[string]int) time.Time {
	result := dateConverter(x.Val(fieldName, headers))
	return result
}

// Converts date from specified field to time.Time
func (x Record) ConvertToTime(fieldName string, headers map[string]int, format ...string) time.Time {
	if len(format) != 0 {
		result, err := time.ParseInLocation(format[0], x.Val(fieldName, headers), time.Local)
		if err != nil {
			config.ErrorF("不能转换为 time.Time: %v", err)
		}

		return result
	} else {
		result, err := time.ParseInLocation(config.TimeFormatDefault, x.Val(fieldName, headers), time.Local)
		if err != nil {
			config.ErrorF("不能转换为 time.Time: %v", err)
		}

		return result
	}
}
