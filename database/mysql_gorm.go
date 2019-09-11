package database

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nocai/infra/consul"
	"github.com/sony/sonyflake"
	"os"
	"time"
)

// 项目关闭时,请Close掉
func NewGormDB(l log.Logger) *gorm.DB {
	username := consul.GetString("database.username")
	password := consul.GetString("database.password")

	address := consul.GetString("database.address")
	database := consul.GetString("database.database")
	_ = level.Info(l).Log("msg", fmt.Sprintf("database[***:***@tcp(%s)/%s]", address, database))

	dialect := consul.GetString("database.dialect")

	maxIdleConns := consul.GetInt("database.maxIdleConns")
	maxOpenConns := consul.GetInt("database.maxOpenConns")
	_ = level.Info(l).Log("msg", fmt.Sprintf("maxIdleConns:%d, maxOpenConns:%d", maxIdleConns, maxOpenConns))

	logMode := consul.GetBool("database.logmode")

	args := `%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local`
	args = fmt.Sprintf(args, username, password, address, database)
	dataBase, err := gorm.Open(dialect, args)
	if err != nil {
		_ = level.Error(l).Log("msg", err)
		os.Exit(-1)
	}

	if err := dataBase.DB().Ping(); err != nil {
		_ = level.Error(l).Log("msg", err)
		os.Exit(-1)
	}

	_ = level.Info(l).Log("msg", fmt.Sprintf("the database[***:***@tcp(%s)/%s] connected", address, database))

	dataBase.DB().SetMaxIdleConns(maxIdleConns)
	dataBase.DB().SetMaxOpenConns(maxOpenConns)

	dataBase.LogMode(logMode)
	//dataBase.SingularTable(true)
	return dataBase
}

type Model struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate 创建前回调
// 用于在ID == 0 时,注入ID
func (m *Model) BeforeCreate() error {
	if m.ID == 0 {
		nextID, err := sonyflake.NewSonyflake(sonyflake.Settings{}).NextID()
		if err != nil {
			return err
		}
		m.ID = nextID
	}
	return nil
}
