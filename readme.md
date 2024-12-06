# guide
this is a module for eav module for go gorm , just support mysql now
go gorm 的eav 模型， 目前只支持mysql

# usage 用法
example 例子

## BasicModelInterface 的實現
``` go
// define connectioin name
// 定義鏈接名字
var testConnectionName string

// define struct
// 定義模型結構體
type UserTest struct {
	EntityId uint64
	Name     string
	Age      uint32
}

// define table of db 
// 定義數據庫名字
type UserTest struct {
	EntityId  uint64
	Name      string
	Age       uint32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// 返回主表名字
func (e *UserTest) GetTableName() string {
	return "user"
}

// 返回字段  ， 會通過反射 reflect 設置 struct 字段
/**
  Name: struct 的字段名字 必須是public 的字段
  IsEav: 是否為eav 字段
   DbType: 跟struct 字段一樣類型
   EavType: 會根據主表名字+ EavType 來確定 eav 的value 表
   Autocreated：  只爲 time.Time 類型處理； 首次創建時會生成
   Autoupdate： 只爲 time.Time 類型處理； 每次保存時會生成
*/
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
// 主表的 primary key
func (e *UserTest) GetPrimaryFieldName() string {
	return "entity_id"
}
```
## 輔助轉換函數和 factory
``` go
// 將 Basictablemodelinterface 轉化為對應類型
func ConvertModelToUserTest(tableModel Basictablemodelinterface) *UserTest {
	model := tableModel.GetModel()

	if m, ok := model.(*UserTest); ok {
		return m
	}

	return &UserTest{}
}

// 獲取 collection
func GetUserTestCollectionFactory(locale string, defaultLocale string) CollectionInterface {
	callback := func() Basictablemodelinterface {
		model := &Basictablemodel{Model: &UserTest{}, Connection: testConnectionName, Locale: locale, DefaultLocale: defaultLocale}
		return model
	}
	return CollectionFactory(callback)
}

// 獲取 Basictablemodelinterface
func GetUserTestFactory(locale string, defaultLocale string) Basictablemodelinterface {
	callback := func() Basictablemodelinterface {
		model := &Basictablemodel{Model: &UserTest{}, Connection: testConnectionName, Locale: locale, DefaultLocale: defaultLocale}
		return model
	}
	return ModelFactory(callback)
}

```

## BasicModelLoadInterface 實現
``` go 
func (e *UserTest) AfterLoad(tablemodel Basictablemodelinterface){
	
}
```

## BasicModelDeleteInterface 實現
``` go 
func (e *UserTest) AfterDelete(tablemodel Basictablemodelinterface){
	
}
```

## BasicModelSaveInterface 實現
``` go 
func (e *UserTest) AfterSave(tablemodel Basictablemodelinterface){
	
}
```