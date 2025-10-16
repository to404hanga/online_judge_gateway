package ioc

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/to404hanga/onlinue_judge_gateway/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func InitDB() *gorm.DB {
	var cfg config.DBConfig
	err := viper.UnmarshalKey(cfg.Key(), &cfg)
	if err != nil {
		log.Panicf("unmarshal db config fail, err: %v", err)
	}

	db, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   cfg.TablePrefix,
			SingularTable: true,
		},
	})
	if err != nil {
		log.Panicf("init db fail, err: %v", err)
	}

	return db
}
