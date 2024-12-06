package core

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type DBConnectionInterface interface {
	Init(adapter string) DBConnectionInterface
	Fetch(sql string) []map[string]interface{}
	FetchOne(sql string) interface{}
	FetchRow(sql string) map[string]interface{}
	Insert(tableName string, values map[string]interface{}) int
	InsertMulti(tableName string, values []map[string]interface{}) DBConnectionInterface
	InsertMultiOnUpdate(tableName string, values []map[string]interface{}) DBConnectionInterface
	Delete(table string, condition string) DBConnectionInterface
	Update(table string, data map[string]interface{}, condition string) DBConnectionInterface
	Expr(sql string, values ...interface{}) string
}

type DBConnection struct {
	Db *gorm.DB
}

func (this *DBConnection) Init(adapter string) DBConnectionInterface {
	db, ok := Db[adapter]
	if !ok {
		return this
	}
	this.Db = db
	return this
}
func (this *DBConnection) Fetch(sql string) []map[string]interface{} {
	var results []map[string]interface{}
	rows, err := this.Db.Raw(sql).Rows()

	if err != nil {
		if this.Db.Error != nil {
			defer func() {
				this.Db.Error = nil
			}()
			panic(this.Db.Error)
		}
		panic(err)
	}
	for rows.Next() {
		// 使用 reflect 包获取列名和数据类型
		cols, err := rows.Columns()
		if err != nil {
			continue
		}

		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}

		err = rows.Scan(valuePtrs...)
		if err != nil {
			continue
		}

		// 将扫描到的值存储到 map 中
		data := map[string]interface{}{}
		for i, col := range cols {
			if values[i] != nil {
				//str := fmt.Sprintf("%v", values[i])
				//data[col] = reflect.ValueOf(values[i]).Interface()
				data[col] = ConvertToString(values[i])
			} else {
				data[col] = nil
			}
		}

		results = append(results, data)
	}
	return results
}

func (this *DBConnection) FetchOne(sql string) interface{} {
	results := this.Fetch(sql)
	if len(results) > 0 {
		for _, result := range results {
			if len(result) > 0 {
				for _, value := range result {
					return value
				}
			}
			break
		}
	}
	return nil
}

func (this *DBConnection) FetchRow(sql string) map[string]interface{} {
	results := this.Fetch(sql)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

func (this *DBConnection) Insert(tableName string, values map[string]interface{}) int {
	var fields []string
	var valuesArr []interface{}
	id := 0
	if len(values) > 0 {
		for key, value := range values {
			fields = append(fields, key)
			valuesArr = append(valuesArr, value)
		}
		sql := "insert into " + tableName + " ("
		valuesStr := "("
		for _, field := range fields {
			sql += field + ","
			valuesStr += "?,"
		}
		valuesStr = strings.TrimSuffix(valuesStr, ",") + ")"
		sql = strings.TrimSuffix(sql, ",") + ") values " + valuesStr
		result := this.Db.Exec(sql, valuesArr...)
		if result.Error != nil {
			fmt.Println("sql error:", result.Error)
		}
		db := result
		if db.Error != nil {
			defer func() {
				db.Error = nil
			}()
			panic(db.Error)
		}

		// 获取最后插入的 ID
		rowsAffected := result.RowsAffected
		if rowsAffected > 0 {
			err := this.Db.Raw("SELECT LAST_INSERT_ID()").Row().Scan(&id)
			if err == nil {
				return id
			} else {
				fmt.Println("sql error:", err)
				if this.Db.Error != nil {
					defer func() {
						this.Db.Error = nil
					}()
					panic(this.Db.Error)
				}
				panic(err)
			}
		}
	}

	return id
}

func (this *DBConnection) InsertMulti(tableName string, values []map[string]interface{}) DBConnectionInterface {
	var fields []string
	var valuesArr []interface{}
	if len(values) > 0 {
		rawValues := values[0]
		for key, _ := range rawValues {
			fields = append(fields, key)
		}
		sql := "insert into " + tableName + " ("
		for _, field := range fields {
			sql += field + ","
		}
		sql = strings.TrimSuffix(sql, ",") + ")"

		valuesStr := ""
		for _, rawValue := range values {
			valuesStr += "("
			for _, field := range fields {
				value := rawValue[field]
				valuesArr = append(valuesArr, value)
				valuesStr += "?,"
			}
			valuesStr = strings.TrimSuffix(valuesStr, ",") + "),"
		}
		valuesStr = strings.TrimSuffix(valuesStr, ",")

		sql = sql + " values " + valuesStr
		db := this.Db.Exec(sql, valuesArr...)
		if db.Error != nil {
			defer func() {
				db.Error = nil
			}()
			panic(db.Error)
		}
	}
	return this
}

func (this *DBConnection) InsertMultiOnUpdate(tableName string, values []map[string]interface{}) DBConnectionInterface {
	var fields []string
	var valuesArr []interface{}
	if len(values) > 0 {
		rawValues := values[0]
		for key, _ := range rawValues {
			fields = append(fields, key)
		}
		sql := "insert into " + tableName + " ("
		duplicateUpdateFieldStr := ""
		for _, field := range fields {
			sql += field + ","
			duplicateUpdateFieldStr += field + "=values(" + field + "), "
		}
		sql = strings.TrimSuffix(sql, ",") + ")"
		duplicateUpdateFieldStr = strings.TrimSuffix(duplicateUpdateFieldStr, ", ")

		valuesStr := ""
		for _, rawValue := range values {
			valuesStr += "("
			for _, field := range fields {
				value := rawValue[field]
				valuesArr = append(valuesArr, value)
				valuesStr += "?,"
			}
			valuesStr = strings.TrimSuffix(valuesStr, ",") + "),"
		}
		valuesStr = strings.TrimSuffix(valuesStr, ",")

		sql += " values " + valuesStr + " ON DUPLICATE KEY UPDATE " + duplicateUpdateFieldStr
		db := this.Db.Exec(sql, valuesArr...)
		if db.Error != nil {
			defer func() {
				db.Error = nil
			}()
			panic(db.Error)
		}
	}
	return this
}

func (this *DBConnection) Delete(table string, condition string) DBConnectionInterface {
	sql := "delete from " + table
	if condition != "" {
		sql += " where " + condition
	}
	db := this.Db.Exec(sql)
	if db.Error != nil {
		defer func() {
			db.Error = nil
		}()
		panic(db.Error)
	}
	return this
}

func (this *DBConnection) Update(table string, data map[string]interface{}, condition string) DBConnectionInterface {
	sql := "update " + table + " set "
	for key, _ := range data {
		sql += " " + key + "=" + "@" + key + ","
	}
	sql = strings.TrimSuffix(sql, ",")
	sql += " where " + condition
	db := this.Db.Exec(sql, data)
	if db.Error != nil {
		defer func() {
			db.Error = nil
		}()
		panic(db.Error)
	}
	return this
}
func (this *DBConnection) Expr(sql string, values ...interface{}) string {

	return this.Db.Raw(sql, values...).ToSQL(func(tx *gorm.DB) *gorm.DB { return tx })
}

var Db map[string]*gorm.DB

func init() {
	Db = make(map[string]*gorm.DB)
}
func AddConnect(connectionName string, db *gorm.DB) {
	Db[connectionName] = db
}
func GetConnection(connectionName string) *gorm.DB {
	db, ok := Db[connectionName]
	if ok {
		return db
	}
	return nil
}
