package model

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	connectDatabase()
	err := DB.AutoMigrate(&User{}, &ChatMsg{}) // TODO: add table structs here
	if err := CreateUser("hdu666", 5); err != nil {
		fmt.Println(err.Error())
	}

	if err != nil {
		logrus.Fatal(err)
	}
}

func connectDatabase() {
	viper.SetConfigName("conf")
	viper.AddConfigPath("./model")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Panic(err)
	}

	loginInfo := viper.GetStringMapString("sql")
	fmt.Println(loginInfo)
	dbArgs := loginInfo["username"] + ":" + loginInfo["password"] +
		"@tcp(localhost:3306)/" + loginInfo["db_name"] + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dbArgs), &gorm.Config{})
	if err != nil {
		logrus.Panic(err)
	}
}
