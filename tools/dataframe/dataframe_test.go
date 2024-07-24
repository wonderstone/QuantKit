package dataframe

import "testing"

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"
)

func TestStream(t *testing.T) {
	firstNameAnswers := []string{"Kevin", "Beth", "Avery", "Peter", "Andy", "Nick", "Bryan", "Brian", "Eric", "Carl"}
	costAnswers := []string{"818", "777", "493", "121", "774", "874", "995", "133", "939", "597"}

	path := "./"
	c := make(chan StreamingRecord)
	go Stream(path, "test/TestData.csv", c)

	i := 0
	for row := range c {
		if row.Val("First Name") != firstNameAnswers[i] {
			t.Error("First name did not match.")
		}
		if row.Val("Cost") != costAnswers[i] {
			t.Error("Cost did not match.")
		}
		i++
	}
}

func TestStreamConvertToInt(t *testing.T) {
	costAnswers := []int64{818, 777, 493, 121, 774, 874, 995, 133, 939, 597}

	path := "./"
	c := make(chan StreamingRecord)
	go Stream(path, "test/TestData.csv", c)

	i := 0
	for row := range c {
		val := row.ConvertToInt("Cost")
		if val != costAnswers[i] {
			t.Error("Could not convert to int64.")
		}
		i++
	}
}

func TestStreamConvertToFloat(t *testing.T) {
	costAnswers := []float64{818.0, 777.0, 493.0, 121.0, 774.0, 874.0, 995.0, 133.0, 939.0, 597.0}

	path := "./"
	c := make(chan StreamingRecord)
	go Stream(path, "test/TestData.csv", c)

	i := 0
	for row := range c {
		val := row.ConvertToFloat("Cost")
		if val != costAnswers[i] {
			t.Error("Could not convert to float64.")
		}
		i++
	}
}

func TestDynamicMetrics(t *testing.T) {
	// Create DataFrame
	columns := []string{"Value"}
	df := CreateNewDataFrame(columns)

	sum := 0.0
	min := 1
	max := 100
	recordedMax := 0.0
	recordedMin := float64(max) + 1.0
	totalRecords := 1_000_000

	for i := 0; i < totalRecords; i++ {
		// Ensures differing values generated on each run.
		rand.Seed(time.Now().UnixNano())
		v := float64(rand.Intn(max-min)+min) + rand.Float64()
		sum = sum + v

		// Add data to DataFrame
		data := []string{fmt.Sprintf("%f", v)}
		df = df.AddRecord(data)

		if v > recordedMax {
			recordedMax = v
		}
		if v < recordedMin {
			recordedMin = v
		}
	}

	dataFrameValue := df.Sum("Value")
	dataFrameAvgValue := math.Round(df.Average("Value")*100) / 100
	dataFrameMaxValue := math.Round(df.Max("Value")*100) / 100
	dataFrameMinValue := math.Round(df.Min("Value")*100) / 100
	avg := math.Round(sum/float64(totalRecords)*100) / 100
	recordedMax = math.Round(recordedMax*100) / 100
	recordedMin = math.Round(recordedMin*100) / 100

	if math.Abs(dataFrameValue-sum) > 0.001 {
		t.Error("Dynamic Metrics: sum float failed", dataFrameValue, sum, math.Abs(dataFrameValue-sum))
	}
	if dataFrameAvgValue != avg {
		t.Error("Dynamic Metrics: average float failed", dataFrameAvgValue, avg)
	}
	if dataFrameMaxValue != recordedMax {
		t.Error("Dynamic Metrics: max value error", dataFrameMaxValue, recordedMax)
	}
	if dataFrameMinValue != recordedMin {
		t.Error("Dynamic Metrics: min value error", dataFrameMinValue, recordedMin)
	}
	if df.CountRecords() != totalRecords {
		t.Error("Dynamic Metrics: count records error", df.CountRecords(), totalRecords)
	}
}

func TestCreateDataFrameCostFloat(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	total := 0.0

	for _, row := range df.FrameRecords {
		total += row.ConvertToFloat("Cost", df.HeaderToIndex)
	}

	if total != 6521.0 {
		t.Error("Cost sum incorrect.")
	}
}

func TestCreateDataFrameCostInt(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	var total int64

	for _, row := range df.FrameRecords {
		total += row.ConvertToInt("Cost", df.HeaderToIndex)
	}

	if total != 6521 {
		t.Error("Cost sum incorrect.")
	}
}

func TestSum(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	if df.Sum("Weight") != 3376.0 || df.Sum("Cost") != 6521.0 {
		t.Error("Just sum error...")
	}
}

func TestAverage(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	if df.Average("Weight") != 337.60 || df.Average("Cost") != 652.10 {
		t.Error("Not your average error...")
	}
}

func TestMax(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	if df.Max("Weight") != 500.0 || df.Max("Cost") != 995.0 {
		t.Error("Error to the max...")
	}
}

func TestMin(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	if df.Min("Weight") != 157.0 || df.Min("Cost") != 121.0 {
		t.Error("Error to the min...")
	}
}

func TestStandardDeviationFunction(t *testing.T) {
	nums := []float64{4.27, 23.45, 34.43, 54.76, 65.90, 234.45}
	stdev := standardDeviation(nums)
	expected := 76.42444976721926
	variance := stdev - expected

	if stdev != expected {
		t.Error(
			fmt.Printf(
				"Standard Deviation calculation error: Expected: %f Output: %f Variance: %f\n", expected, stdev,
				variance,
			),
		)
	}
}

func TestStandardDeviationMethodPass(t *testing.T) {
	// Create DataFrame
	columns := []string{"RID", "Value"}
	df := CreateNewDataFrame(columns)

	for i := 0; i < 1000; i++ {
		val := strconv.Itoa(i)
		df = df.AddRecord([]string{"RID-" + val, val})
	}

	stdev, err := df.StandardDeviation("Value")
	if err != nil {
		t.Error("Test should have passed without any string to float conversion errors.")
	}

	expected := 288.6749902572095
	variance := stdev - expected

	if stdev != expected {
		t.Error(
			fmt.Printf(
				"Standard Deviation calculation error: Expected: %f Output: %f Variance: %f\n", expected, stdev,
				variance,
			),
		)
	}
}

func TestStandardDeviationMethodFail(t *testing.T) {
	// Create DataFrame
	columns := []string{"RID", "Value"}
	df := CreateNewDataFrame(columns)

	for i := 0; i < 1000; i++ {
		// Insert row with value that cannot be converted to float64.
		if i == 500 {
			df = df.AddRecord([]string{"RID-" + "500", "5x0x0x"})
		}
		val := strconv.Itoa(i)
		df = df.AddRecord([]string{"RID-" + val, val})
	}

	_, err := df.StandardDeviation("Value")
	if err == nil {
		t.Error("Test should have failed.")
	}
}

func TestFilteredCount(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	dfFil := df.Filtered("Last Name", "Fultz", "Wiedmann")

	if df.CountRecords() != 10 || dfFil.CountRecords() != 5 {
		t.Error("Filtered count incorrect.")
	}
}

func TestFilteredCheck(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	dfFil := df.Filtered("Last Name", "Fultz", "Wiedmann")

	for _, row := range dfFil.FrameRecords {
		if row.Val("Last Name", dfFil.HeaderToIndex) != "Fultz" && row.Val(
			"Last Name", dfFil.HeaderToIndex,
		) != "Wiedmann" {
			t.Error("Invalid parameter found in Filtered DataFrame.")
		}
	}
}

// Ensures changes made in the original dataframe are not also made in a filtered dataframe.
func TestFilteredChangeToOriginal(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	dfFil := df.Filtered("Last Name", "Fultz", "Wiedmann")

	for _, row := range df.FrameRecords {
		if row.Val("RID", df.HeaderToIndex) == "2" {
			row.Update("Last Name", "Bethany", df.HeaderToIndex)
		}
		if row.Val("RID", df.HeaderToIndex) == "5" {
			row.Update("Last Name", "Andyanne", df.HeaderToIndex)
		}
	}

	// Ensure row was actually updated in the original frame.
	for _, row := range df.FrameRecords {
		if row.Val("RID", df.HeaderToIndex) == "2" && row.Val("Last Name", df.HeaderToIndex) != "Bethany" {
			t.Error("Row 2 last name not changed in original frame.")
		}
		if row.Val("RID", df.HeaderToIndex) == "5" && row.Val("Last Name", df.HeaderToIndex) != "Andyanne" {
			t.Error("Row 5 last name not changed in original frame.")
		}
	}

	// Check rows in filtered dataframe were not also updated.
	for _, row := range dfFil.FrameRecords {
		if row.Val("RID", df.HeaderToIndex) == "2" && row.Val("Last Name", df.HeaderToIndex) != "Fultz" {
			t.Error("Row 2 in filtered dataframe was incorrectly updated with original.")
		}
		if row.Val("RID", df.HeaderToIndex) == "5" && row.Val("Last Name", df.HeaderToIndex) != "Wiedmann" {
			t.Error("Row 5 in filtered dataframe was incorrectly updated with original.")
		}
	}
}

func TestGreaterThanOrEqualTo(t *testing.T) {
	path := "./"
	value := float64(597)
	df := CreateDataFrame(path, "test/TestData.csv")
	df, err := df.GreaterThanOrEqualTo("Cost", value)
	if err != nil {
		t.Error("Greater Than Or Equal To: This should not have failed...")
	}

	if df.CountRecords() != 7 {
		t.Error("Greater Than Or Equal To: Record count is not correct.")
	}

	ids := []string{"1", "2", "5", "6", "7", "9", "10"}
	foundIds := df.Unique("RID")

	for i, id := range foundIds {
		if id != ids[i] {
			t.Error("Greater Than Or Equal To: Records do not match.")
		}
	}
}

func TestLessThanOrEqualTo(t *testing.T) {
	path := "./"
	value := float64(436)
	df := CreateDataFrame(path, "test/TestData.csv")
	df, err := df.LessThanOrEqualTo("Weight", value)
	if err != nil {
		t.Error("Less Than Or Equal To: This should not have failed...")
	}

	if df.CountRecords() != 7 {
		t.Error("Less Than Or Equal To: Record count is not correct.")
	}

	ids := []string{"1", "2", "4", "5", "6", "8", "9"}
	foundIds := df.Unique("RID")

	for i, id := range foundIds {
		if id != ids[i] {
			t.Error("Less Than Or Equal To: Records do not match.")
		}
	}
}

func TestExcludeCount(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	dfExcl := df.Exclude("Last Name", "Fultz", "Wiedmann")

	if df.CountRecords() != 10 || dfExcl.CountRecords() != 5 {
		t.Error("Excluded count is incorrect.")
	}
}

func TestExcludeCheck(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	dfExcl := df.Exclude("Last Name", "Fultz", "Wiedmann")

	for _, row := range dfExcl.FrameRecords {
		if row.Val("Last Name", dfExcl.HeaderToIndex) == "Fultz" || row.Val(
			"Last Name", dfExcl.HeaderToIndex,
		) == "Wiedmann" {
			t.Error("Excluded parameter found in DataFrame.")
		}
	}
}

func TestFilteredAfterCount(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	dfFil := df.FilteredAfter("Date", "2022-01-08")

	if df.CountRecords() != 10 || dfFil.CountRecords() != 2 {
		t.Error("Filtered After count incorrect.")
	}
}

func TestFilteredAfterCountExcelFormat(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestDataDateFormat.csv")
	dfFil := df.FilteredAfter("Date", "2022-01-08")

	if df.CountRecords() != 10 || dfFil.CountRecords() != 2 {
		t.Error("Filtered After Excel Format count incorrect.")
	}
}

func TestFilteredBeforeCount(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	dfFil := df.FilteredBefore("Date", "2022-01-08")

	if df.CountRecords() != 10 || dfFil.CountRecords() != 7 {
		t.Error("Filtered Before count incorrect.")
	}
}

func TestFilteredBeforeCountExcelFormat(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestDataDateFormat.csv")
	dfFil := df.FilteredBefore("Date", "2022-01-08")

	if df.CountRecords() != 10 || dfFil.CountRecords() != 7 {
		t.Error("Filtered Before Excel Format count incorrect.")
	}
}

func TestFilteredBetweenCount(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	dfFil := df.FilteredBetween("Date", "2022-01-02", "2022-01-09")

	if df.CountRecords() != 10 || dfFil.CountRecords() != 6 {
		t.Error("Filtered Between count incorrect.")
	}
}

func TestFilteredBetweenExcelFormat(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestDataDateFormat.csv")
	dfFil := df.FilteredBetween("Date", "2022-01-02", "2022-01-09")

	if df.CountRecords() != 10 || dfFil.CountRecords() != 6 {
		t.Error("Filtered Between Excel Format count incorrect.")
	}
}

func TestRecordCheck(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	var id string
	var date string
	var cost string
	var weight string
	var firstName string
	var lastName string

	for _, row := range df.FrameRecords {
		if row.Val("RID", df.HeaderToIndex) == "5" {
			id = row.Val("RID", df.HeaderToIndex)
			date = row.Val("Date", df.HeaderToIndex)
			cost = row.Val("Cost", df.HeaderToIndex)
			weight = row.Val("Weight", df.HeaderToIndex)
			firstName = row.Val("First Name", df.HeaderToIndex)
			lastName = row.Val("Last Name", df.HeaderToIndex)
		}
	}

	if id != "5" {
		t.Error("RID failed")
	} else if date != "2022-01-05" {
		t.Error("Date failed")
	} else if cost != "774" {
		t.Error("Cost failed")
	} else if weight != "415" {
		t.Error("Weight failed")
	} else if firstName != "Andy" {
		t.Error("First Name failed")
	} else if lastName != "Wiedmann" {
		t.Error("Last Name failed")
	}
}

func TestRecordCheckPanic(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	for _, row := range df.FrameRecords {
		defer func() { recover() }()

		row.Val("Your Name Here", df.HeaderToIndex)

		// Never reaches here if `OtherFunctionThatPanics` panics.
		t.Errorf("The row.Val() method should have panicked.")
	}
}

func TestAddRecord(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	newData := [6]string{"11", "2022-06-23", "101", "500", "Ben", "Benison"}
	df = df.AddRecord(newData[:])

	if df.CountRecords() != 11 {
		t.Error("Add Record: Count does not match.")
	}

	for _, row := range df.FrameRecords {
		if row.Val("RID", df.HeaderToIndex) == "11" {
			if row.Val("Date", df.HeaderToIndex) != "2022-06-23" {
				t.Error("Add Record: date failed")
			}
			if row.Val("Cost", df.HeaderToIndex) != "101" {
				t.Error("Add Record: cost failed")
			}
			if row.Val("Weight", df.HeaderToIndex) != "500" {
				t.Error("Add Record: weight failed")
			}
			if row.Val("First Name", df.HeaderToIndex) != "Ben" {
				t.Error("Add Record: first name failed")
			}
			if row.Val("Last Name", df.HeaderToIndex) != "Benison" {
				t.Error("Add Record: last name failed")
			}
		}
	}
}

func TestByteOrderMark(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestDataCommaSeparatedValue.csv")
	dfUtf := CreateDataFrame(path, "test/TestData.csv")

	dfTotal := 0.0
	for _, row := range df.FrameRecords {
		dfTotal += row.ConvertToFloat("RID", df.HeaderToIndex)
	}

	dfUtfTotal := 0.0
	for _, row := range dfUtf.FrameRecords {
		dfUtfTotal += row.ConvertToFloat("RID", dfUtf.HeaderToIndex)
	}

	if dfTotal != 55.0 || dfUtfTotal != 55.0 {
		t.Error("Byte Order Mark conversion error")
	}
}
func TestKeepColumns(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	columns := [3]string{"First Name", "Last Name", "Weight"}
	df = df.KeepColumns(columns[:])

	if df.HeaderToIndex["First Name"] != 0 || df.HeaderToIndex["Last Name"] != 1 || df.HeaderToIndex["Weight"] != 2 || len(df.HeaderToIndex) > 3 {
		t.Error("Keep Columns failed")
	}
}

func TestRemoveColumnsMultiple(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	df = df.RemoveColumns("RID", "Cost", "First Name")

	if df.HeaderToIndex["Date"] != 0 || df.HeaderToIndex["Weight"] != 1 || df.HeaderToIndex["Last Name"] != 2 || len(df.HeaderToIndex) > 3 {
		t.Error("Remove Multiple Columns failed")
	}
}

func TestRemoveColumnsSingle(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	df = df.RemoveColumns("First Name")

	if df.HeaderToIndex["RID"] != 0 || df.HeaderToIndex["Date"] != 1 || df.HeaderToIndex["Cost"] != 2 || df.HeaderToIndex["Weight"] != 3 || df.HeaderToIndex["Last Name"] != 4 || len(df.HeaderToIndex) > 5 {
		t.Error("Remove Single Column failed")
	}
}

func TestDateConverterStandardFormat(t *testing.T) {
	var s interface{} = dateConverter("2022-01-31")
	if _, ok := s.(time.Time); ok != true {
		t.Error("Date Converter Standard Format Failed")
	}
}

func TestDateConverterExcelFormatDoubleDigit(t *testing.T) {
	var s interface{} = dateConverter("01/31/2022")
	if _, ok := s.(time.Time); ok != true {
		t.Error("Date Converter Excel Format Failed")
	}
}

func TestDateConverterExcelFormatSingleMonthDigit(t *testing.T) {
	var s interface{} = dateConverter("1/31/2022")
	if _, ok := s.(time.Time); ok != true {
		t.Error("Date Converter Excel Format Failed")
	}
}

func TestDateConverterExcelFormatSingleDayDigit(t *testing.T) {
	var s interface{} = dateConverter("01/1/2022")
	if _, ok := s.(time.Time); ok != true {
		t.Error("Date Converter Excel Format Failed")
	}
}

func TestDateConverterExcelFormatSingleDigit(t *testing.T) {
	var s interface{} = dateConverter("1/1/2022")
	if _, ok := s.(time.Time); ok != true {
		t.Error("Date Converter Excel Format Failed")
	}
}

func TestDateConverterExcelFormatDoubleYearDigit(t *testing.T) {
	var s interface{} = dateConverter("01/31/22")
	if _, ok := s.(time.Time); ok != true {
		t.Error("Date Converter Excel Format Failed")
	}
}

func TestNewField(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	df.NewField("Middle Name")

	if df.HeaderToIndex["Middle Name"] != 6 {
		fmt.Println(df.HeaderToIndex)
		t.Error("New field column not added in proper position.")
	}

	for _, row := range df.FrameRecords {
		if row.Val("Middle Name", df.HeaderToIndex) != "" {
			t.Error("Value in New Field is not set to nil")
		}
	}
}

func TestUnique(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	names := df.Unique("Last Name")

	if len(names) != 7 {
		t.Error("Unique slice error.")
	}
}

func TestUpdate(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	for _, row := range df.FrameRecords {
		if row.Val("First Name", df.HeaderToIndex) == "Avery" && row.Val("Last Name", df.HeaderToIndex) == "Fultz" {
			row.Update("Weight", "30", df.HeaderToIndex)
		}
	}

	for _, row := range df.FrameRecords {
		if row.Val("First Name", df.HeaderToIndex) == "Avery" && row.Val("Last Name", df.HeaderToIndex) == "Fultz" {
			if row.Val("Weight", df.HeaderToIndex) != "30" {
				t.Error("DoOrderUpdate row failed.")
			}
		}
	}
}

func TestUpdatePanic(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	for _, row := range df.FrameRecords {
		if row.Val("First Name", df.HeaderToIndex) == "Avery" && row.Val("Last Name", df.HeaderToIndex) == "Fultz" {
			defer func() { recover() }()

			row.Update("Your Name Here", "30", df.HeaderToIndex)

			t.Errorf("Method should have panicked.")
		}
	}
}

func TestMergeFramesAllColumns(t *testing.T) {
	path := "./"

	// Prep left frame
	df := CreateDataFrame(path, "test/TestData.csv")
	// newData := [6]string{"11", "2022-06-27", "5467", "9586", "Cassandra", "SchmaSandra"}
	// df = df.AddRecord(newData[:])

	// Prep right frame
	dfRight := CreateDataFrame(path, "test/TestMergeData.csv")

	// Merge
	df.Merge(&dfRight, "RID")

	fmt.Println(df)
	if df.CountRecords() != 11 {
		t.Error("Merge: record count error.")
	}

	m := make(map[string][]string)
	m["2"] = []string{"RICHLAND", "WA", "99354"}
	m["4"] = []string{"VAN BUREN", "AR", "72956"}
	m["6"] = []string{"FISHERS", "NY", "14453"}
	m["10"] = []string{"JEFFERSON CITY", "MO", "65109"}
	m["11"] = []string{"", "", ""}

	df.ViewColumns()

	for _, row := range df.FrameRecords {
		if val, ok := m[row.Val("RID", df.HeaderToIndex)]; ok {
			for i, v := range val {
				switch i {
				case 0:
					if row.Val("City", df.HeaderToIndex) != v {
						t.Error("Merge: city error.")
					}
				case 1:
					if row.Val("State", df.HeaderToIndex) != v {
						t.Error("Merge: state error.")
					}
				case 2:
					if row.Val("Postal Code", df.HeaderToIndex) != v {
						t.Error("Merge: postal code error.")
					}
				}
			}
		}
	}
}

func TestMergeFramesSpecifiedColumns(t *testing.T) {
	path := "./"

	// Prep left frame
	df := CreateDataFrame(path, "test/TestData.csv")
	newData := [6]string{"11", "2022-06-27", "5467", "9586", "Cassandra", "SchmaSandra"}
	df = df.AddRecord(newData[:])

	// Prep right frame
	dfRight := CreateDataFrame(path, "test/TestMergeData.csv")

	// Merge
	df.Merge(&dfRight, "RID", "City", "Postal Code")

	if df.CountRecords() != 11 {
		t.Error("Merge: record count error.")
	}

	m := make(map[string][]string)
	m["2"] = []string{"RICHLAND", "99354"}
	m["4"] = []string{"VAN BUREN", "72956"}
	m["6"] = []string{"FISHERS", "14453"}
	m["10"] = []string{"JEFFERSON CITY", "65109"}
	m["11"] = []string{"", ""}

	for _, row := range df.FrameRecords {
		if val, ok := m[row.Val("RID", df.HeaderToIndex)]; ok {
			for i, v := range val {
				switch i {
				case 0:
					if row.Val("City", df.HeaderToIndex) != v {
						t.Error("Merge: city error.")
					}
				case 1:
					if row.Val("Postal Code", df.HeaderToIndex) != v {
						t.Error("Merge: postal code error.")
					}
				}
			}
		}
	}
}

func TestInnerMerge(t *testing.T) {
	path := "./"

	// Prep left frame
	df := CreateDataFrame(path, "test/TestData.csv")

	// Prep right frame
	dfRight := CreateDataFrame(path, "test/TestInnerMergeData.csv")

	// Merge
	df = df.InnerMerge(&dfRight, "RID")

	if df.CountRecords() != 5 {
		t.Error("Inner Merge: record count error.")
	}

	columns := []string{"RID", "Date", "Cost", "Weight", "First Name", "Last Name", "City", "State", "Postal Code"}

	data := make([][]string, 5)
	data[0] = []string{"4", "2022-01-04", "121", "196", "Peter", "Wiedmann", "VAN BUREN", "AR", "72956"}
	data[1] = []string{"5", "2022-01-05", "774", "415", "Andy", "Wiedmann", "TAUNTON", "MA", "2780"}
	data[2] = []string{"7", "2022-01-07", "995", "500", "Bryan", "Curtis", "GOLDSBORO", "NC", "27530"}
	data[3] = []string{"9", "2022-01-09", "939", "157", "Eric", "Petruska", "PHOENIX", "AZ", "85024"}
	data[4] = []string{"10", "2022-01-10", "597", "475", "Carl", "Carlson", "JEFFERSON CITY", "MO", "65109"}

	for i, row := range df.FrameRecords {
		if len(row.Data) != len(data[i]) {
			t.Error("Inner Merge: Column count does not match.")
		}
		for i2, col := range columns {
			val := row.Val(col, df.HeaderToIndex)
			if val != data[i][i2] {
				t.Error("Inner Merge: Data results to not match what is expected.")
			}
		}
	}
}

func TestInnerMergeLeftFrameDuplicates(t *testing.T) {
	path := "./"

	// Prep left frame
	df := CreateDataFrame(path, "test/TestDataInnerDuplicate.csv")

	// Prep right frame
	dfRight := CreateDataFrame(path, "test/TestInnerMergeData.csv")

	// Merge
	df = df.InnerMerge(&dfRight, "RID")

	if df.CountRecords() != 6 {
		t.Error("Inner Merge: record count error.")
	}

	columns := []string{"RID", "Date", "Cost", "Weight", "First Name", "Last Name", "City", "State", "Postal Code"}

	data := make([][]string, 6)
	data[0] = []string{"4", "2022-01-04", "121", "196", "Peter", "Wiedmann", "VAN BUREN", "AR", "72956"}
	data[1] = []string{"5", "2022-01-05", "774", "415", "Andy", "Wiedmann", "TAUNTON", "MA", "2780"}
	data[2] = []string{"7", "2022-01-07", "995", "500", "Bryan", "Curtis", "GOLDSBORO", "NC", "27530"}
	data[3] = []string{"9", "2022-01-09", "939", "157", "Eric", "Petruska", "PHOENIX", "AZ", "85024"}
	data[4] = []string{"9", "2022-01-09", "12345", "6789", "Eric", "Petruska", "PHOENIX", "AZ", "85024"}
	data[5] = []string{"10", "2022-01-10", "597", "475", "Carl", "Carlson", "JEFFERSON CITY", "MO", "65109"}

	for i, row := range df.FrameRecords {
		if len(row.Data) != len(data[i]) {
			t.Error("Inner Merge: Column count does not match.")
		}
		for i2, col := range columns {
			val := row.Val(col, df.HeaderToIndex)
			if val != data[i][i2] {
				t.Error("Inner Merge: Data results to not match what is expected.")
			}
		}
	}
}

func TestConcatFrames(t *testing.T) {
	path := "./"
	dfOne := CreateDataFrame(path, "test/TestData.csv")
	df := CreateDataFrame(path, "test/TestDataConcat.csv")

	lastNames := [20]string{
		"Fultz",
		"Fultz",
		"Fultz",
		"Wiedmann",
		"Wiedmann",
		"Wilfong",
		"Curtis",
		"Wenck",
		"Petruska",
		"Carlson",
		"Benny",
		"Kenny",
		"McCarlson",
		"Jeffery",
		"Stephenson",
		"Patrickman",
		"Briarson",
		"Ericson",
		"Asherton",
		"Highman",
	}

	dfOne, err := dfOne.ConcatFrames(&df)
	if err != nil {
		t.Error("Concat Frames: ", err)
	}
	var totalCost int64
	var totalWeight int64

	for i, row := range dfOne.FrameRecords {
		if row.Val("Last Name", dfOne.HeaderToIndex) != lastNames[i] {
			t.Error("Concat Frames Failed: Last Names")
		}
		totalCost += row.ConvertToInt("Cost", dfOne.HeaderToIndex)
		totalWeight += row.ConvertToInt("Weight", dfOne.HeaderToIndex)
	}

	if totalCost != 7100 || totalWeight != 3821 {
		t.Error("Concat Frames Failed: Values")
	}

	if dfOne.CountRecords() != 20 {
		t.Error("Concat Frames Failed: Row Count")
	}
}

func TestConcatFramesAddress(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	df2 := CreateDataFrame(path, "test/TestDataConcat.csv")

	df3, err := df.ConcatFrames(&df2)
	if err != nil {
		t.Error(err)
	}

	if &df == &df3 || &df2 == &df3 {
		t.Error("ConcatFrames did not create a truly decoupled new dataframe")
	}
	if df3.CountRecords() != 20 {
		t.Error("ConcatFrames did not properly append")
	}
}

func TestConcatFramesColumnCount(t *testing.T) {
	path := "./"
	dfOne := CreateDataFrame(path, "test/TestData.csv")
	columns := []string{"one", "two", "three"}
	dfTwo := CreateNewDataFrame(columns)

	dfOne, err := dfOne.ConcatFrames(&dfTwo)
	if err == nil {
		t.Error("Concat Frames Did Not Fail --> ", err)
	}
}

func TestConcatFramesColumnOrder(t *testing.T) {
	path := "./"
	dfOne := CreateDataFrame(path, "test/TestData.csv")
	columns := []string{
		"RID",
		"Date",
		"Cost",
		"Weight",
		"Last Name",
		"First Name",
	}
	dfTwo := CreateNewDataFrame(columns)

	dfOne, err := dfOne.ConcatFrames(&dfTwo)
	if err == nil {
		t.Error("Concat Frames Did Not Fail --> ", err)
	}
}

// Ensures once a new filtered DataFrame is created, if records are updated in the original
// it will not affect the records in the newly created filtered version.
func TestCopiedFrame(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	df2 := df.Filtered("Last Name", "Wiedmann")

	// DoOrderUpdate data in original frame.
	for _, row := range df.FrameRecords {
		if row.Val("First Name", df.HeaderToIndex) == "Peter" && row.Val("Last Name", df.HeaderToIndex) == "Wiedmann" {
			row.Update("Last Name", "New Last Name", df.HeaderToIndex)
		}
	}

	// Check value did not change in newly copied frame.
	for _, row := range df2.FrameRecords {
		if row.Val("RID", df2.HeaderToIndex) == "4" {
			if row.Val("First Name", df2.HeaderToIndex) != "Peter" || row.Val(
				"Last Name", df2.HeaderToIndex,
			) != "Wiedmann" {
				t.Error("Copied Frame: name appears to have changed in second frame.")
			}
		}
	}
}

func TestSaveDataFrame(t *testing.T) {
	path := "./"
	// df := CreateDataFrame(path, "test/TestData.csv")
	read := make(chan StreamingRecord)

	go Stream(path, "test/TestData.csv", read)
	WriteStream(
		path, "test/Testing.csv", read, []string{"Date", "Cost"}, func(record StreamingRecord) []string {
			data := []string{record.Data[1], record.Data[2]}
			return data
		},
	)
}

func TestAssortment(t *testing.T) {
	path := "./"

	// Concatenate Frames
	dfOne := CreateDataFrame(path, "test/TestData.csv")
	df := CreateDataFrame(path, "test/TestDataConcat.csv")
	df, err := df.ConcatFrames(&dfOne)
	if err != nil {
		log.Fatal("Concat Frames: ", err)
	}

	// Add Records
	newData := [6]string{"21", "2022-01-01", "200", "585", "Tommy", "Thompson"}
	df = df.AddRecord(newData[:])
	newDataTwo := [6]string{"22", "2022-01-31", "687", "948", "Sarah", "McSarahson"}
	df = df.AddRecord(newDataTwo[:])

	if df.CountRecords() != 22 {
		t.Error("Assortment: concat count incorrect.")
	}

	df = df.Exclude("Last Name", "Fultz", "Highman", "Stephenson")

	if df.CountRecords() != 17 {
		t.Error("Assortment: excluded count incorrect.")
	}

	df = df.FilteredAfter("Date", "2022-01-08")

	if df.CountRecords() != 4 {
		t.Error("Assortment: filtered after count incorrect.")
	}

	lastNames := df.Unique("Last Name")
	checkLastNames := [4]string{"Petruska", "Carlson", "Asherton", "McSarahson"}

	if len(lastNames) != 4 {
		t.Error("Assortment: last name count failed")
	}

	for _, name := range lastNames {
		var status bool
		for _, cName := range checkLastNames {
			if name == cName {
				status = true
			}
		}
		if status != true {
			t.Error("Assortment: last name not found.")
		}
	}

}

func TestCopy(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	df2 := df.Copy()

	for _, row := range df2.FrameRecords {
		if row.Val("First Name", df2.HeaderToIndex) == "Bryan" && row.Val("Last Name", df2.HeaderToIndex) == "Curtis" {
			row.Update("First Name", "Brian", df2.HeaderToIndex)
		}
		if row.Val("First Name", df2.HeaderToIndex) == "Carl" && row.Val("Last Name", df2.HeaderToIndex) == "Carlson" {
			row.Update("First Name", "McCarlson", df2.HeaderToIndex)
		}
	}

	// Test original frame did not change.
	for _, row := range df.FrameRecords {
		if row.Val("Last Name", df.HeaderToIndex) == "Curtis" {
			if row.Val("First Name", df.HeaderToIndex) != "Bryan" {
				t.Error("First Name in original frame is not correct.")
			}
		}
		if row.Val("Last Name", df.HeaderToIndex) == "Carlson" {
			if row.Val("First Name", df.HeaderToIndex) != "Carl" {
				t.Error("First Name in original frame is not correct.")
			}
		}
	}

	// Test copied frame contains changes.
	for _, row := range df2.FrameRecords {
		if row.Val("Last Name", df2.HeaderToIndex) == "Curtis" {
			if row.Val("First Name", df2.HeaderToIndex) != "Brian" {
				t.Error("First Name in copied frame is not correct.")
			}
		}
		if row.Val("Last Name", df2.HeaderToIndex) == "Carlson" {
			if row.Val("First Name", df2.HeaderToIndex) != "McCarlson" {
				t.Error("First Name in copied frame is not correct.")
			}
		}
	}
}

func TestCopyAddress(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")
	df2 := df.Copy()

	if &df == &df2 {
		t.Error("Copy did not create a truly decoupled copy.")
	}
}

func TestColumns(t *testing.T) {
	path := "./"
	requiredColumns := []string{
		"RID",
		"Date",
		"Cost",
		"Weight",
		"First Name",
		"Last Name",
	}
	df := CreateDataFrame(path, "test/TestData.csv")
	foundColumns := df.Columns()

	if len(foundColumns) != 6 {
		t.Error("Length of found columns does not match")
	}

	for i := 0; i < len(requiredColumns); i++ {
		if foundColumns[i] != requiredColumns[i] {
			t.Error("Order of found columns does not match")
		}
	}
}

func TestAutoCount(t *testing.T) {
	columns := []string{"id", "number", "value"}
	df := CreateNewDataFrame(columns)

	for i := 0; i < 1_000; i++ {
		val := float64(i + 1)
		sq := val * val
		data := []string{
			strconv.Itoa(i),
			fmt.Sprintf("%f", val),
			fmt.Sprintf("%f", sq),
		}
		df = df.AddRecord(data)
	}

	if df.CountRecords() != 1_000 {
		t.Error("Test Auto: count is not 1,000,000")
	}
}

func TestAutoSum(t *testing.T) {
	columns := []string{"id", "number", "value"}
	df := CreateNewDataFrame(columns)

	for i := 0; i < 1_000; i++ {
		val := float64(i + 1)
		sq := val * val
		data := []string{
			strconv.Itoa(i),
			fmt.Sprintf("%f", val),
			fmt.Sprintf("%f", sq),
		}
		df = df.AddRecord(data)
	}

	if df.Sum("value") != 333_833_500.0 {
		t.Error("Test Auto: sum is not correct")
	}
}

func TestLoadFrames(t *testing.T) {
	filePath := "./"
	files := []string{
		"test/TestData.csv",
		"test/TestDataCommaSeparatedValue.csv",
		"test/TestDataConcat.csv",
		"test/TestDataDateFormat.csv",
		"test/TestMergeData.csv",
	}

	results, err := LoadFrames(filePath, files)
	if err != nil {
		log.Fatal(err)
	}

	dfTd := results[0]
	dfComma := results[1]
	dfConcat := results[2]
	dfDate := results[3]
	dfMerge := results[4]

	if dfTd.CountRecords() != 10 || dfTd.Sum("Weight") != 3376.0 || len(dfTd.Columns()) != 6 {
		t.Error("LoadFrames: TestData.csv is not correct")
	}
	if dfComma.CountRecords() != 10 || dfComma.Sum("Cost") != 6521.0 || len(dfComma.Columns()) != 6 {
		t.Error("LoadFrames: TestDataCommaSeparatedValue.csv is not correct")
	}
	if dfConcat.CountRecords() != 10 || dfConcat.Sum("Weight") != 445.0 || len(dfConcat.Columns()) != 6 {
		t.Error("LoadFrames: TestDataConcat.csv is not correct")
	}
	if dfDate.CountRecords() != 10 || dfDate.Average("Cost") != 652.1 || len(dfDate.Columns()) != 6 {
		t.Error("LoadFrames: TestDataDateFormat.csv is not correct")
	}
	if dfMerge.CountRecords() != 10 || dfMerge.Sum("Postal Code") != 495735.0 || len(dfMerge.Columns()) != 4 {
		t.Error("LoadFrames: TestMergeData.csv is not correct")
	}

	dfFilterTest := dfTd.Filtered("Last Name", "Fultz")
	if dfTd.CountRecords() == dfFilterTest.CountRecords() {
		t.Error("LoadFrame: variable referencing map value")
	}
}

func TestLoadFramesError(t *testing.T) {
	filePath := "./"
	files := []string{"test/TestData.csv"}

	_, err := LoadFrames(filePath, files)
	if err == nil {
		t.Error("LoadFrames did not fail as expected")
	}
}

func TestRename(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	err := df.Rename("Weight", "Total Weight")
	if err != nil {
		t.Error(err)
	}

	for _, row := range df.FrameRecords {
		if row.Val("First Name", df.HeaderToIndex) == "Andy" && row.Val("Last Name", df.HeaderToIndex) == "Wiedmann" {
			row.Update("Total Weight", "1000", df.HeaderToIndex)
		}
	}

	for _, row := range df.FrameRecords {
		if row.Val("First Name", df.HeaderToIndex) == "Andy" && row.Val("Last Name", df.HeaderToIndex) == "Wiedmann" {
			if row.Val("Total Weight", df.HeaderToIndex) != "1000" {
				t.Error("Value in new column did not update correctly")
			}
		}
	}

	foundColumns := []string{}
	newColumnStatus := false
	for k, _ := range df.HeaderToIndex {
		foundColumns = append(foundColumns, k)
		if k == "Total Weight" {
			newColumnStatus = true
		}
	}

	if newColumnStatus != true {
		t.Error("New column was not found")
	}
	if len(foundColumns) != 6 {
		t.Error("Wrong number of columns found")
	}
}

func TestRenameOriginalNotFound(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	err := df.Rename("The Weight", "Total Weight")
	if err == nil {
		t.Error(err)
	}
}

func TestRenameDuplicate(t *testing.T) {
	path := "./"
	df := CreateDataFrame(path, "test/TestData.csv")

	err := df.Rename("Weight", "Cost")
	if err == nil {
		t.Error(err)
	}
}
