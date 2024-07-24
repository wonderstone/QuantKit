package factor

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

// 根据数据字段查询索引表以获取包含该字段的表名
func GetTableNameByField(db *gorm.DB, indexTableName string, dataField string) (string, error) {
	var result struct {
		TableName string
	}

	err := db.Table(indexTableName).Select("table_name").Where("factor = ?", dataField).Take(&result).Error
	if err != nil {
		return "", err
	}

	return result.TableName, nil
}

// 根据提供的表名、合约代码、数据字段和时间范围查询数据
func GetDataByFieldAndContract(
	db *gorm.DB, tableName string, dataField string, startTime time.Time, endTime time.Time,
) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := db.Table(tableName).Select("date", dataField).Where(
		"date BETWEEN ? AND ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"),
	)
	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// 封装查询：首先通过索引表确定数据所在的表，然后查询数据
func GetDataListByContractCodeAndField(
	db *gorm.DB, indexTableName string, dataField string, startTime time.Time, endTime time.Time,
) ([]map[string]interface{}, error) {
	tableName, err := GetTableNameByField(db, indexTableName, dataField)
	if err != nil {
		return nil, err
	}

	if tableName == "" {
		return nil, errors.New("因子字段未在任何表中找到")
	}

	return GetDataByFieldAndContract(db, tableName, dataField, startTime, endTime)
}
