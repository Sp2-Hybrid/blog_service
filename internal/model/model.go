package model

import (
	"blog_service/global"
	"blog_service/pkg/setting"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type Model struct {
	ID         uint32 `gorm:"primary_key" json:"id"`
	CreatedBy  string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
	CreatedOn  uint32 `json:"created_on"`
	ModifiedOn uint32 `json:"modified_on"`
	DeletedOn  uint32 `json:"deleted_on"`
	IsDel      uint8  `json:"is_del"`
}

func NewDBEngine(databaseSetting *setting.DatabaseSettingS) (*gorm.DB, error) {
	db, err := gorm.Open(databaseSetting.DBType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=%t&loc=Local",
		databaseSetting.UserName,
		databaseSetting.Password,
		databaseSetting.Host,
		databaseSetting.DBName,
		databaseSetting.Charset,
		databaseSetting.ParseTime,
	))
	if err != nil {
		return nil, err
	}

	if global.ServerSetting.RunMode == "debug" {
		db.LogMode(true)
	}
	db.SingularTable(true)
	db.DB().SetMaxIdleConns(databaseSetting.MaxIdleConns)
	db.DB().SetMaxOpenConns(databaseSetting.MaxOpenConns)

	return db, nil
}

func updateTimeStampForCreateCallback(scope *gorm.Scope){
	if !scope.HasError(){
		nowTime := time.Now().Unix()
		if createTimeField, ok := scope.FieldByName("CreatedOn");ok{
			if createTimeField.IsBlank{
				_ = createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("ModifiedOn"); ok{
			if modifyTimeField.IsBlank{
				_ = modifyTimeField.Set(nowTime)
			}
		}
	}
}

func updateTimeStampForUpdateCallback(scope *gorm.Scope){

}

func deleteCallback(scope *gorm.Scope){

}

func addExtraSpaceIfExist(str string)string{

}