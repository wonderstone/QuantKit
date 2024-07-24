// dailyFactor 日频因子数据源数据
// 1. 从sqlite中读取数据
// 2. 提供数据源参数
package indicator

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/container/queue"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/factor"
	"gorm.io/gorm"
)

type ksdtf struct {
	Date time.Time
	Val  float64
}

// DailyFactor is the moving average indicator
type KSDDailyFactor struct {
	calculator formula.Calculator
	Name       string
	instID     string
	source     string
	table      string
	isEmpty    bool

	cachedTime *time.Time

	curVal *ksdtf
	db     *gorm.DB

	DQ *queue.Queue[ksdtf]
}

// ! 这个可以删掉
func (m *KSDDailyFactor) SetInstID(instID string) {
	m.instID = instID
}

func (m *KSDDailyFactor) SetCalculator(calculator formula.Calculator) {
	m.calculator = calculator
}

func (m *KSDDailyFactor) DoInit(conf config.Formula) {
	m.Name = conf.Name
	// ! 没添加m.instID
	m.instID = conf.InstID

	m.DQ = queue.New[ksdtf](1000)

	if handler, ok := conf.Param["Handler"]; ok {
		if handler != "sqlite" {
			config.ErrorF("日频因子%s只支持sqlite数据源", m.Name)
		}

		if source, ok := conf.Param["Source"]; ok {
			m.source = source
		} else {
			config.ErrorF("日频因子%s未指定数据源", m.Name)
		}
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(m.source, m.instID+".db")), &gorm.Config{})
	if err != nil {
		config.ErrorF("无法连接数据库")
	}

	m.db = db

	table, err := factor.GetTableNameByField(
		m.db, "t_index", m.Name,
	)
	if err != nil {
		config.ErrorF("无法获取因子表名")
	}

	m.table = table
}

func (m *KSDDailyFactor) loadData(tm time.Time) {
	var results []map[string]any

	query := m.db.Table(m.table).Select("Date", m.Name).Where(
		"date >= ? limit 1000", tm.Format("2006-01-02"),
	)

	err := query.Find(&results).Error
	if err != nil {
		config.ErrorF("无法获取因子数据")
	}

	if len(results) == 0 {
		m.isEmpty = true
		return
	}

	for _, result := range results {
		d := result["date"].(time.Time)
		d = d.Add(16 * time.Hour)

		val := result[m.Name].(float64)
		m.DQ.EnqueueWithDequeue(ksdtf{d, val})
	}

	if tail, ok := m.DQ.PeekTail(); ok {
		m.cachedTime = &tail.Date
	}
}
// ! 新接口  返回值为string
func (m *KSDDailyFactor) DoCalculate(tm time.Time, row dataframe.RecordFunc) string{
	if m.isEmpty {
		return ""
	}

	for {
		if m.DQ.Len() == 0 {
			m.cachedTime = nil
			if m.isEmpty {
				return ""
			}

			m.loadData(tm)
			continue
		}

		postVal := m.DQ.MustGet(0)
		if postVal.Date.After(tm) {
			break
		}

		if curVal, ok := m.DQ.Dequeue(); ok {
			m.curVal = &curVal
		}
	}
	// ! 增加返回值
	if m.curVal != nil {
		// 推测已经无用
		// row.Update(m.Name, fmt.Sprintf("%.2f", m.curVal.Val), m.calculator.GetColumns())
		return fmt.Sprintf("%.2f", m.curVal.Val)
	}else{
		return ""
	}
	
	

}

func (m *KSDDailyFactor) DoReset() {
	db, err := gorm.Open(sqlite.Open(filepath.Join(m.source, m.instID+".db")), &gorm.Config{})
	if err != nil {
		config.ErrorF("无法连接数据库")
	}

	m.db = db

	table, err := factor.GetTableNameByField(
		m.db, "t_index", m.Name,
	)
	if err != nil {
		config.ErrorF("无法获取因子表名")
	}

	m.table = table
}

func init() {
	formula.RegisterNewFormula(new(KSDDailyFactor), "KSDDailyFactor")
}
