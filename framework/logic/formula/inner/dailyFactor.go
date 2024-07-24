// dailyFactor 日频因子数据源数据
// 1. 从sqlite中读取数据
// 2. 提供数据源参数
package indicator

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"sync"

	"github.com/glebarez/sqlite"
	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/factor"
	"gorm.io/gorm"
)

type tf struct {
	Date string
	Val  float64
}

// DailyFactor is info
type DailyFactor struct {
	Name       string
	InstID     string
	Factor     string
	Mode       string
	Source     string
	Table      string
	IsEmpty    bool
	Data       *sync.Map
	lock       sync.Mutex
	firstDate  time.Time
	// cachedTime *time.Time

	curVal *tf
	db     *gorm.DB
}



func (m *DailyFactor) DoInit(conf config.Formula) {
	m.Name = conf.Name
	m.InstID = conf.InstID
	m.Mode = conf.Param["Mode"]

	m.lock = sync.Mutex{}
	data := sync.Map{}
	m.Data = &data
	// m.Data = sync.Map{}

	if factor, ok := conf.Param["Factor"]; ok {
		m.Factor = factor
	} else {
		config.ErrorF("日频因子%s未指定因子名称", m.Name)
	}

	if handler, ok := conf.Param["Handler"]; ok {
		if handler != "sqlite" {
			config.ErrorF("日频因子%s只支持sqlite数据源", m.Name)
		}

		if source, ok := conf.Param["Source"]; ok {
			m.Source = source
		} else {
			config.ErrorF("日频因子%s未指定数据源", m.Name)
		}
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(m.Source, m.InstID+".db")), &gorm.Config{})
	if err != nil {
		config.ErrorF("无法连接数据库")
	}

	m.db = db

	table, err := factor.GetTableNameByField(
		m.db, "t_index", m.Factor,
	)
	if err != nil {
		config.ErrorF("无法获取因子表名")
	}

	m.Table = table

	// read in data from sqlite
	m.lock.Lock()
	defer m.lock.Unlock()
	var results []map[string]any
	query := m.db.Table(m.Table).Select("Date", m.Factor)
	err = query.Find(&results).Error
	if err != nil {
		config.ErrorF("无法获取因子数据")
	}
	if len(results) == 0 {
		m.IsEmpty = true
		return
	}

	m.firstDate = results[0]["Date"].(time.Time)
	firstDate := m.firstDate.Format("2006-01-02")
	
	m.curVal = &tf{
		Date: firstDate,
		Val:  results[0][m.Factor].(float64),
	}
	// 假定外部日线数据时间标签为入库时间 且第二天可用
	for _, result := range results {
		d := result["Date"].(time.Time).Format("2006-01-02")
		val := result[m.Factor].(float64)
		m.Data.Store(d, val)
	}

}

func (m *DailyFactor) loadData(tm time.Time) {

	if strings.ToLower(m.Mode) == "rt" {
		var results []map[string]any

		query := m.db.Table(m.Table).Select("Date", m.Factor).Where(
			"date >= ? limit 1000", tm.Format("2006-01-02"),
		)

		err := query.Find(&results).Error
		if err != nil {
			config.ErrorF("无法获取因子数据")
		}

		for _, result := range results {
			d := result["date"]
			// d = d.Add(16 * time.Hour)

			val := result[m.Factor].(float64)
			m.Data.Store(d, val)
		}

	}

}
 
// func get a time.Time last date
func getLastdate(tm time.Time) time.Time {
	tm = tm.AddDate(0, 0, -1)
	return tm
}

func (m *DailyFactor) DoCalculate(tm time.Time, row dataframe.RecordFunc) string {

	m.loadData(tm)
	if m.IsEmpty {
		return ""
	}
	if tm.Before(m.firstDate) {
		return ""
	}
	// m.cachedTime = &tm
	// v, ok:=m.data.Load(tm)
	// iter the m.data to find the tm if not use getLastdate and find again until find the tm
	// if not find the tm return the last date
	for {
		tmp := tm.Format("2006-01-02")
		// if tmp == "2012-06-29"{
		// 	fmt.Println("2012-06-29")
		// }
		
		
		v, ok := m.Data.Load(tmp)
		
		if ok {
			m.curVal = &tf{
				Date: tm.Format("2006-01-02"),
				Val:  v.(float64),
			}
		} else {
			tm = getLastdate(tm)
			continue
		}
		break
	}
	// fmt.Sprintf 会自动四舍五入
	return fmt.Sprintf("%.4f", m.curVal.Val)
}

func (m *DailyFactor) DoReset() {
	db, err := gorm.Open(sqlite.Open(filepath.Join(m.Source, m.InstID+".db")), &gorm.Config{})
	if err != nil {
		config.ErrorF("无法连接数据库")
	}

	m.db = db

	table, err := factor.GetTableNameByField(
		m.db, "t_index", m.Factor,
	)
	if err != nil {
		config.ErrorF("无法获取因子表名")
	}

	m.Table = table
}

func init() {
	formula.RegisterNewFormula(new(DailyFactor), "DailyFactor")
}

func NewDailyFactor(Name string, instID ,mode,source,fct string) *DailyFactor {

	//instID "600000.XSHG.CS"
	//mode "train"
	//source "/Users/wonderstone/Desktop/QT/DMTF-dir/download/factor"
	//factor "pe"

	db, err := gorm.Open(sqlite.Open(filepath.Join(source, instID+".db")), &gorm.Config{})
	if err != nil {
		config.ErrorF("无法连接数据库")
	}

	table, err := factor.GetTableNameByField(
		db, "t_index", fct,
	)
	if err != nil {
		config.ErrorF("无法获取因子表名")
	}

	var results []map[string]any
	query := db.Table(table).Select("Date", fct)
	err = query.Find(&results).Error
	if err != nil {
		config.ErrorF("无法获取因子数据")
	}

	fd := results[0]["date"].(time.Time).Format("2006-01-02")
	curVal := &tf{
		Date: fd,
		Val:  results[0][fct].(float64),
	}
	// 假定外部日线数据时间标签为入库时间 且第二天可用
	data := sync.Map{}
	for _, result := range results {
		d := result["date"].(time.Time).Format("2006-01-02")
		// d = d.Add(16 * time.Hour)

		val := result[fct].(float64)
		data.Store(d, val)
		// fmt.Println(d, val)
	}

	return &DailyFactor{
		Name:   Name,
		InstID: instID,
		Mode:   mode,
		Source: source,
		Table:  table,
		Factor: fct,
		Data:   &data,
		curVal: curVal,
		db:     db,
	}
}
