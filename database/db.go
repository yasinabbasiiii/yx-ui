package database

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path"

	"x-ui/config"
	"x-ui/database/model"

	"x-ui/xray"

	"gorm.io/driver/mysql" //samyar
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var db2 *gorm.DB
var db3 *gorm.DB

var initializers = []func() error{
	initUser,
	initInbound,
	initOutbound,
	initSetting,
	initInboundClientIps,
	initClientTraffic,
	initClientTrafficDetails,
}

func initUser() error {
	err := db2.AutoMigrate(&model.User{})
	if err != nil {
		return err
	}
	var count int64
	err = db2.Model(&model.User{}).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		user := &model.User{
			Username:    "admin",
			Password:    "admin",
			LoginSecret: "",
		}
		return db2.Create(user).Error
	}
	return nil
}

func initInbound() error {
	return db2.AutoMigrate(&model.Inbound{})
}

func initOutbound() error {
	return db2.AutoMigrate(&model.OutboundTraffics{})
}

func initSetting() error {
	return db2.AutoMigrate(&model.Setting{})
}

func initInboundClientIps() error {
	return db2.AutoMigrate(&model.InboundClientIps{})
}

func initClientTraffic() error {
	return db2.AutoMigrate(&xray.ClientTraffic{})
}

// Samyar
func initClientTrafficDetails() error {
	return db3.AutoMigrate(&xray.ClientTrafficDetails{})
}
func InitDB(dbPath string) error {
	dir := path.Dir(dbPath)
	err := os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		return err
	}

	var gormLogger logger.Interface

	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	c := &gorm.Config{
		Logger: gormLogger,
	}
	// اتصال به MySQL
	dsn := "yas:Yas2566*7425@tcp(db.ir107.ir:3306)/x_ui"
	db, err = gorm.Open(mysql.Open(dsn), c)
	if err != nil {
		//logger.Error = "connect to db failed, trying to connect to db1"
		dsn := "yas:Yas2566*7425@tcp(db1.ir107.ir:3306)/x_ui"
		db, err = gorm.Open(mysql.Open(dsn), c)
		if err != nil {
			//logger.Error("connect to db1 failed!")
			return err
		}
	}
	// اتصال به MySQL
	db3, err = gorm.Open(mysql.Open("yas:Yas2566*7425@tcp(db.ir107.ir:3306)/x_ui_2"), c)
	if err != nil {
		//logger.Error = "connect to db failed, trying to connect to db1"
		db3, err = gorm.Open(mysql.Open("yas:Yas2566*7425@tcp(db1.ir107.ir:3306)/x_ui_2"), c)
		if err != nil {
			//logger.Error("connect to db1 failed!")
			return err
		}
	}

	// اتصال به SQLite
	db2, err = gorm.Open(sqlite.Open(dbPath), c)
	if err != nil {
		return err
	}
	for _, initialize := range initializers {
		if err := initialize(); err != nil {
			return err
		}
	}

	return nil
}

// اتصال به MySQL
func GetDB() *gorm.DB {
	return db
}

// اتصال به MySQL
func GetDB3() *gorm.DB {
	return db3
}

// اتصال به SQLite
func GetDB2() *gorm.DB {
	return db2
}
func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func IsSQLiteDB(file io.ReaderAt) (bool, error) {
	signature := []byte("SQLite format 3\x00")
	buf := make([]byte, len(signature))
	_, err := file.ReadAt(buf, 0)
	if err != nil {
		return false, err
	}
	return bytes.Equal(buf, signature), nil
}

func Checkpoint() error {
	// Update WAL
	err := db2.Exec("PRAGMA wal_checkpoint;").Error
	if err != nil {
		return err
	}
	return nil
}
