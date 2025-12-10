package core

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var testConnectionName string

type UserTest struct {
	EntityId  uint64
	Name      string
	Age       uint32
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (e *UserTest) GetTableName() string {
	return "user"
}

func (e *UserTest) GetTableFields() map[string]Field {
	fields := map[string]Field{
		"entity_id":  {Name: "EntityId", IsEav: false, DbType: "uint64"},
		"name":       {Name: "Name", IsEav: true, DbType: "string", EavType: "varchar"},
		"age":        {Name: "Age", IsEav: false, DbType: "uint32"},
		"created_at": {Name: "CreatedAt", IsEav: false, DbType: "time.Time", Autocreate: true},
		"updated_at": {Name: "UpdatedAt", IsEav: false, DbType: "time.Time", Autocreate: true, Autoupdate: true},
	}
	return fields
}
func (e *UserTest) GetPrimaryFieldName() string {
	return "entity_id"
}

func (e *UserTest) BeforeSave(tablemodel Basictablemodelinterface) {
	if tablemodel.GetData("name") == "error" {
		panic("cannot save name if error")
	}
}

func ConvertModelToUserTest(tableModel Basictablemodelinterface) *UserTest {
	model := tableModel.GetModel()

	if m, ok := model.(*UserTest); ok {
		return m
	}

	return &UserTest{}
}

func GetUserTestCollectionFactory(locale string, defaultLocale string) CollectionInterface {
	callback := func() Basictablemodelinterface {
		model := &Basictablemodel{Model: &UserTest{}, Connection: testConnectionName, Locale: locale, DefaultLocale: defaultLocale}
		return model
	}
	return CollectionFactory(callback)
}
func GetUserTestFactory(locale string, defaultLocale string) Basictablemodelinterface {
	callback := func() Basictablemodelinterface {
		model := &Basictablemodel{Model: &UserTest{}, Connection: testConnectionName, Locale: locale, DefaultLocale: defaultLocale}
		return model
	}
	return ModelFactory(callback)
}

const testDBName = "testdb"

var testDB *gorm.DB

func TestMain(m *testing.M) {
	// 数据库连接字符串（不包含数据库名）
	dsn := "root:bc123123@tcp(127.0.0.1:3308)/?charset=utf8mb4&parseTime=True&loc=UTC"

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 检查数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// 创建测试数据库
	createDatabaseSQL := "CREATE DATABASE IF NOT EXISTS " + testDBName
	if _, err := db.Exec(createDatabaseSQL); err != nil {
		log.Fatalf("failed to create database: %v", err)
	}

	// 更新 dsn 包含新创建的数据库
	dsn = "root:bc123123@tcp(127.0.0.1:3308)/" + testDBName + "?charset=utf8mb4&parseTime=True&loc=UTC"
	migratedsn := "mysql://" + dsn

	// 应用数据库迁移
	migrationSource := "file://../migrations"
	migration, err := migrate.New(migrationSource, migratedsn)
	if err != nil {
		log.Fatalf("failed to create migration instance: %v", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to apply migrations: %v", err)
	}
	testConnectionName = "test"
	InitDB(dsn, testConnectionName)
	defer CloseDb(testConnectionName)

	// 运行测试
	exitCode := m.Run()

	// 清理测试数据库
	if err := migration.Down(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to downgrade migrations: %v", err)
	}
	os.Exit(exitCode)
}

func TestCreateUser(t *testing.T) {
	assert := assert.New(t)

	userModel := GetUserTestFactory("en-US", "en-US")
	userModel.SetData("name", "John Doe").SetData("age", 22).Save()
	user := ConvertModelToUserTest(userModel)
	assert.NotEqual(user.EntityId, uint64(0))
	assert.Equal("John Doe", user.Name)
	assert.True(user.Age == 22)
	assert.False(user.UpdatedAt.IsZero(), "updated at is zero")
	assert.False(user.CreatedAt.IsZero(), "updated at is zero")
}

func TestCreateErrorUser(t *testing.T) {
	assert := assert.New(t)

	userModel := GetUserTestFactory("en-US", "en-US")
	userModel.SetData("name", "error").SetData("age", 22).Save()
	user := ConvertModelToUserTest(userModel)
	assert.Equal(user.EntityId, uint64(0))
	error := userModel.GetLastError()
	assert.NotNil(error)
	assert.EqualError(error, "cannot save name if error")
}

func TestGetUserByID(t *testing.T) {
	assert := assert.New(t)

	// 插入一个用户
	userModel := GetUserTestFactory("en-US", "en-US")
	userModel.SetData("name", "John Doe22").SetData("age", 23).Save()
	user := ConvertModelToUserTest(userModel)
	assert.NotNil(user.EntityId)

	// 获取用户
	userModel2 := GetUserTestFactory("en-US", "en-US")
	userModel2.LoadById(user.EntityId)
	user2 := ConvertModelToUserTest(userModel2)
	assert.NotNil(user2.EntityId)
	assert.Equal(user.Name, user2.Name)
	assert.Equal(user.Age, user2.Age)
}

func TestUpdateUser(t *testing.T) {
	assert := assert.New(t)

	// 插入一个用户
	userModel := GetUserTestFactory("en-US", "en-US")
	userModel.SetData("name", "John Doe22").SetData("age", 23).Save()
	user := ConvertModelToUserTest(userModel)
	assert.NotNil(user.EntityId)

	// 更新用户
	user.Name = "Alice Smith"
	userModel.SetData("name", "Jake Doe").SetData("age", 22).Save()

	// 获取用户
	userModel2 := GetUserTestFactory("zh-CN", "en-US")
	userModel2.LoadById(user.EntityId)
	user2 := ConvertModelToUserTest(userModel2)
	assert.NotNil(user2.EntityId)
	assert.Equal(user.Name, user2.Name)
	assert.Equal(user.Age, user2.Age)
	userModel2.SetData("name", "傑克 doe").SetData("age", 12).Save()

	// 获取用户 en-US
	userModel3 := GetUserTestFactory("en-US", "en-US")
	userModel3.LoadById(user.EntityId)
	user3 := ConvertModelToUserTest(userModel3)
	assert.NotNil(user3.EntityId)
	assert.Equal(user.Name, user3.Name)
	assert.NotEqual(user.Age, user3.Age)
	assert.NotEqual(user2.Name, user3.Name)
	assert.Equal(user2.Age, user3.Age)

	// 获取用户 zh-CN
	userModel4 := GetUserTestFactory("zh-CN", "en-US")
	userModel4.LoadById(user.EntityId)
	user4 := ConvertModelToUserTest(userModel4)
	assert.NotNil(user4.EntityId)
	assert.NotEqual(user.Name, user4.Name)
	assert.NotEqual(user.Age, user4.Age)
	assert.Equal(user2.Name, user4.Name)
	assert.Equal(user2.Age, user4.Age)
}

func TestDeleteUser(t *testing.T) {
	assert := assert.New(t)

	// 插入一个用户
	userModel := GetUserTestFactory("en-US", "en-US")
	userModel.SetData("name", "John Doe22").SetData("age", 23).Save()
	user := ConvertModelToUserTest(userModel)
	assert.NotNil(user.EntityId)
	id := user.EntityId

	// 删除用户
	userModel.Delete()

	// 检查用户是否被删除
	// 获取用户
	userModel2 := GetUserTestFactory("zh-CN", "en-US")
	userModel2.LoadById(id)
	user2 := ConvertModelToUserTest(userModel2)
	assert.NotNil(id)
	assert.NotEqual(user.Name, user2.Name)
	assert.NotEqual(user.Age, user2.Age)
}

func TestListUsers(t *testing.T) {
	assert := assert.New(t)

	userModelCollection := GetUserTestCollectionFactory("en-US", "en-US")
	for _, user := range userModelCollection.GetElems() {
		user.Delete()
	}

	// 插入多个用户
	users := []map[string]interface{}{
		{"name": "User1", "age": 11, "zh_name": "中文 user1"},
		{"name": "User2", "age": "12", "zh_name": "中文 user2"},
		{"name": "User3", "age": "13", "zh_name": "中文 user3"},
	}
	for _, user := range users {
		userModel := GetUserTestFactory("en-US", "en-US")
		userModel.SetData("name", user["name"]).SetData("age", user["age"]).Save()
		dbuser := ConvertModelToUserTest(userModel)
		userModel2 := GetUserTestFactory("zh-CN", "en-US").LoadById(dbuser.EntityId)
		userModel2.SetData("name", user["zh_name"]).Save()
	}

	// 列出所有用户
	userModelCollection2 := GetUserTestCollectionFactory("en-US", "en-US")
	dbUserModels := userModelCollection2.GetElems()
	assert.Len(dbUserModels, 3)
	index := 0
	for _, dbuser := range dbUserModels {
		user := users[index]
		index += 1
		assert.True(user["name"] == dbuser.GetData("name"))
		assert.True(ConvertToInt32(user["age"]) == ConvertToInt32(dbuser.GetData("age")))
	}

	// 列出所有用户
	userModelCollection3 := GetUserTestCollectionFactory("zh-CN", "en-US")
	dbUserModels3 := userModelCollection3.GetElems()
	assert.Len(dbUserModels3, 3)
	index3 := 0
	for _, dbuser := range dbUserModels3 {
		user := users[index3]
		index3 += 1
		assert.True(user["zh_name"] == dbuser.GetData("name"))
		assert.True(ConvertToInt32(user["age"]) == ConvertToInt32(dbuser.GetData("age")))
	}

}

func TestUtils(t *testing.T) {
	assert := assert.New(t)
	// 測試時間
	/**
	return map[string]string{
		"UTC":            "(00:00) Universal Time",
		"Asia/Shanghai":  "(+08:00) Beijing, Shanghai",
		"Asia/Hong_Kong": "(+08:00) Hong Kong",
	}
	*/

	timestr := "2024-12-03 12:00:00"
	assert.Equal(FormatTimeFromUTCToLocale(timestr, "UTC"), "2024-12-03 12:00:00")
	assert.Equal(FormatTimeFromUTCToLocale(timestr, "Asia/Shanghai"), "2024-12-03 20:00:00")
	assert.Equal(FormatTimeFromLocaleToUTC(timestr, "UTC"), "2024-12-03 12:00:00")
	assert.Equal(FormatTimeFromLocaleToUTC(timestr, "Asia/Shanghai"), "2024-12-03 04:00:00")
	time := ConvertToTime(timestr)
	assert.Equal(time.IsZero(), false)
	assert.Equal(ConvertToTimeString(time), "2024-12-03 12:00:00")
	assert.Equal(ConvertToTimeStringFromUTCToLocale(time, "Asia/Shanghai"), "2024-12-03 20:00:00")
	assert.Equal(ConvertToTimeStringFromLocaleToUTC(time, "Asia/Shanghai"), "2024-12-03 04:00:00")

	timestr = ""
	assert.Equal(FormatTimeFromUTCToLocale(timestr, "UTC"), "")
	assert.Equal(FormatTimeFromUTCToLocale(timestr, "Asia/Shanghai"), "")
	assert.Equal(FormatTimeFromLocaleToUTC(timestr, "UTC"), "")
	assert.Equal(FormatTimeFromLocaleToUTC(timestr, "Asia/Shanghai"), "")
	time = ConvertToTime(timestr)
	assert.Equal(time.IsZero(), true)
	assert.Equal(ConvertToTimeString(time), "")
	assert.Equal(ConvertToTimeStringFromUTCToLocale(time, "Asia/Shanghai"), "")
	assert.Equal(ConvertToTimeStringFromLocaleToUTC(time, "Asia/Shanghai"), "")

}

func testTransaction(t *testing.T) {
	assert := assert.New(t)
	userModelCollection := GetUserTestCollectionFactory("en-US", "en-US")
	for _, user := range userModelCollection.GetElems() {
		user.Delete()
	}

	db := GetConnection("default")
	// 开始事务
	db.Transaction(func(tx *gorm.DB) error {
		// 插入多个用户
		users := []map[string]interface{}{
			{"name": "User1-tran", "age": 11, "zh_name": "中文 user1"},
			{"name": "User2-tran", "age": "12", "zh_name": "中文 user2"},
			{"name": "User3-tran", "age": "13", "zh_name": "中文 user3"},
		}
		for index, user := range users {
			tx.Transaction(func(tx *gorm.DB) error {
				userModel := GetUserTestFactory("en-US", "en-US")
				userModel.GetConnection().SetDb(tx)
				userModel.SetData("name", user["name"]).SetData("age", user["age"]).Save()
				dbuser := ConvertModelToUserTest(userModel)
				userModel2 := GetUserTestFactory("zh-CN", "en-US")
				userModel2.GetConnection().SetDb(tx)
				userModel2.LoadById(dbuser.EntityId)
				userModel2.SetData("name", user["zh_name"]).Save()
				return nil
			})
			if index == 2 {
				return errors.New("test transaction error")
			}
		}
		return nil
	})

	userModelCollection3 := GetUserTestCollectionFactory("zh-CN", "en-US")
	dbUserModels3 := userModelCollection3.GetElems()
	assert.Len(dbUserModels3, 0)
}
