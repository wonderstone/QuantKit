package recorder

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/wonderstone/QuantKit/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Sqlite[T any] struct {
	db   *gorm.DB
	flag sync.Once
	data chan any

	enableTransaction bool
}

func (s *Sqlite[T]) QueryRecord(query ...WithQuery) []any {
	op := NewQuery(query...)

	if !op.sqlMode {
		config.ErrorF("SqliteRecorder目前只支持SQL查询")
		return nil
	}

	var result []T
	s.db.Raw(op.sql).Scan(&result)

	var res []any
	for _, r := range result {
		res = append(res, r)
	}
	return res
}

func NewSqliteRecorder[T any](option ...WithOption) *Sqlite[T] {
	op := NewOp(option...)

	if !strings.HasSuffix(op.file, ".db") {
		op.file += ".db"
	}

	// 文件夹不存在则创建
	dir := filepath.Dir(op.file)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			config.ErrorF("创建文件夹失败: %v", err.Error())
		}
	}

	if !op.plusMode {
		if _, err := os.Stat(op.file); !os.IsNotExist(err) {
			err := os.Remove(op.file)
			if err != nil {
				config.ErrorF("删除文件失败: %v", err.Error())
			}
		}
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
		},
	)

	db, err := gorm.Open(
		sqlite.Open(op.file), &gorm.Config{
			Logger:                 newLogger,
			PrepareStmt:            true,
			SkipDefaultTransaction: true,
			// CreateBatchSize:        1000,
		},
	)

	if err != nil {
		config.ErrorF("创建数据库失败: %s", err.Error())
	}

	return &Sqlite[T]{db: db, data: make(chan any, 100), enableTransaction: op.transaction}
}

func (s *Sqlite[T]) RecordChan() error {
	if s.enableTransaction {
		return s.db.Transaction(
			func(tx *gorm.DB) error {
				// tx := s.db.Session(&gorm.Session{PrepareStmt: true})
				for d := range s.data {
					s.flag.Do(
						func() {
							err := tx.AutoMigrate(d)
							if err != nil {
								config.ErrorF("创建表失败: %s", err.Error())
							}
						},
					)
					tx.Save(d)
				}

				return nil
			},
		)
	} else {
		cache := make([]*T, 0, 1000)
		for d := range s.data {
			s.flag.Do(
				func() {
					err := s.db.AutoMigrate(d)
					if err != nil {
						config.ErrorF("创建表失败: %s", err.Error())
					}
				},
			)

			cache = append(cache, d.(*T))

			if len(cache) < 1000 {
				continue
			}

			tx := s.db.Model(d).Clauses(clause.OnConflict{UpdateAll: true}).Create(cache)
			if tx.Error != nil {
				config.ErrorF("批量写入失败 err %v", tx.Error)
			}

			cache = make([]*T, 0, 1000)
		}

		if len(cache) > 0 {
			tx := s.db.Model(cache[0]).Clauses(clause.OnConflict{UpdateAll: true}).Create(cache)
			if tx.Error != nil {
				config.ErrorF("批量写入失败 err %v", tx.Error)
			}
		}
	}

	return nil
}

func (s *Sqlite[T]) Read(data any) error {
	s.db.Find(data)
	return nil
}

func (s *Sqlite[T]) GetChannel() chan any {
	return s.data
}

func (s *Sqlite[T]) Release() {
	close(s.data)
}
