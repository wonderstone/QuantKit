package recorder

import (
	// "encoding/csv"
	// "fmt"

	// "os"
	"testing"
	"time"
	// "github.com/gocarina/gocsv"
	// "github.com/jinzhu/gorm"
)

func TestCsv(t *testing.T) {
	// 
	opt:= []WithOption{
		WithFilePath("test.csv"),
		WithPlusMode(),
		WithTransaction(),
	}


	recorder := NewCsvRecorder(opt...)

	type Test struct {
		Name string    `csv:"name"`
		Age  int       `csv:"age"`
		Addr string    `csv:"addr"`
		Time time.Time `csv:"time"`
	}

	go func() {
	err := recorder.RecordChan()
	if err != nil {
	panic(err)
	}
	}()

	for i := 0; i < 1000; i++ {
		recorder.GetChannel() <- &Test{
			Name: "John,Doe",
			Age:  i,
			Addr: "123",
			Time: time.Now(),
		}
	}

	close(recorder.GetChannel())
}

// func TestSqlite(t *testing.T) {
// recorder := NewSqliteRecorder("test.db")

// type Test struct {
// gorm.Model
// Name string
// Age  int
// Addr string
// Time time.Time
// }

// err := recorder.db.AutoMigrate(&Test{})
// if err != nil {
// return
// }

// go func() {
// err := recorder.RecordChan()
// if err != nil {
// panic(err)
// }
// 	}()

// for i := 0; i < 1000; i++ {
// recorder.GetChannel() <- &Test{
// Name: "John,Doe",
// Age:  i,
// Addr: "123",
// Time: time.Now(),
// }
// 	}

// close(recorder.GetChannel())
// }

// func BenchmarkCsv2(b *testing.B) {
// 	type Test struct {
// 		Name string    `csv:"name"`
// 		Age  int       `csv:"age"`
// 		Addr string    `csv:"addr"`
// 		Time time.Time `csv:"time"`
// 	}

// 	c := make(chan any)

// 	file, _ := os.Create("test.csv")
// 	writer := csv.NewWriter(file)
// 	go func() {
// 		// defer wg.Done()
// 		err := gocsv.MarshalChan(c, writer)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}()

// 	for i := 0; i < b.N; i++ {
// 		c <- &Test{
// 			Name: "John,Doe",
// 			Age:  i,
// 			Addr: "123",
// 			Time: time.Now(),
// 		}
// 	}

// 	close(c)
// }

// func TestMemory(t *testing.T) {
// 	type Test struct {
// 		Name string    `csv:"name"`
// 		Age  int       `csv:"age"`
// 		Addr string    `csv:"addr"`
// 		Time time.Time `csv:"time"`
// 	}

// 	recorder := NewMemoryRecorder[Test]()

// 	go func() {
// 	err := recorder.RecordChan(recorder.GetChannel())
// 	if err != nil {
// 	panic(err)
// 	}
// 	}()

// 	for i := 0; i < 1000; i++ {
// 		recorder.GetChannel() <- Test{
// 			Name: "John,Doe",
// 			Age:  i,
// 			Addr: "123",
// 			Time: time.Now(),
// 		}
// 	}

// 	close(recorder.GetChannel())

// 	for _, v := range recorder.GetRecord() {
// 		fmt.Println(v)
// 	}
// }
