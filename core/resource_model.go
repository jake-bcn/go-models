package core

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type BasictableResourceInterface interface {
	Initialize(model Basictablemodelinterface, adapter string) Basictablemodelinterface
	SetData(field string, value interface{}) BasictableResourceInterface
	GetData(field string) interface{}
	SetOriginData(field string, value interface{}) BasictableResourceInterface
	GetOriginData(field string) interface{}
	LoadDbData(data map[string]interface{}) BasictableResourceInterface
	Reset() BasictableResourceInterface
	Save() BasictableResourceInterface
	LoadByField(field string, value interface{}) BasictableResourceInterface
	Delete() BasictableResourceInterface
	GetConnection() DBConnectionInterface
	GetEavAsTable() string
}

type basictableResource struct {
	Model      Basictablemodelinterface
	Data       map[string]interface{}
	OriginData map[string]interface{}
	Connection DBConnectionInterface
}

func (e *basictableResource) Initialize(model Basictablemodelinterface, adapter string) Basictablemodelinterface {
	e.Model = model
	e.Connection = &DBConnection{}
	e.Connection.Init(adapter)
	e.Reset()
	return model
}

func (e *basictableResource) Convert(field string, value interface{}) interface{} {
	handleValue := value
	if handleValue == nil {
		return nil
	}
	fields := e.Model.GetTableFields()
	if def, ok := fields[field]; ok {
		switch def.DbType {
		case "int":
			handleValue = ConvertToInt(value)

		case "int64":
			handleValue = ConvertToInt64(value)

		case "int32":
			handleValue = ConvertToInt32(value)

		case "int16":
			handleValue = ConvertToInt16(value)

		case "uint":
			handleValue = ConvertToUint(value)

		case "uint8":
			handleValue = ConvertToUint8(value)

		case "uint64":
			handleValue = ConvertToUint64(value)

		case "uint32":
			handleValue = ConvertToUint32(value)

		case "uint16":
			handleValue = ConvertToUint16(value)

		case "string":
			handleValue = ConvertToString(value)

		case "float64":
			handleValue = ConvertToFloat64(value)

		case "float32":
			handleValue = ConvertToFloat32(value)

		case "time.Time":
			handleValue = ConvertToTimeString(value)
			if handleValue == "" {
				handleValue = nil
			}

		}
	}
	return handleValue
}
func (e *basictableResource) GetFieldDefByName(field string) (*Field, bool) {
	fields := e.Model.GetTableFields()
	for key, fieldDef := range fields {
		if strings.EqualFold(key, field) {
			return &fieldDef, true
		}
	}
	return nil, false
}
func (e *basictableResource) SetData(field string, value interface{}) BasictableResourceInterface {
	field = strings.ToLower(field)
	handleValue := e.Convert(field, value)
	if modelField, ok := e.GetFieldDefByName(field); ok {
		// 通过反射设置结构体变量的值
		// time.Time  // TODO: aa
		if modelField.DbType == "time.Time" {
			t := ConvertToTime(handleValue)
			updateField(e.Model.GetModel(), modelField.Name, t)
		} else {
			updateField(e.Model.GetModel(), modelField.Name, handleValue)
		}

	}

	e.Data[field] = handleValue
	return e
}

func (e *basictableResource) GetData(field string) interface{} {
	field = strings.ToLower(field)
	value := e.Data[field]
	return value
}

func (e *basictableResource) SetOriginData(field string, value interface{}) BasictableResourceInterface {
	field = strings.ToLower(field)
	handleValue := e.Convert(field, value)
	e.OriginData[field] = handleValue
	return e
}

func (e *basictableResource) GetOriginData(field string) interface{} {
	field = strings.ToLower(field)
	value := e.OriginData[field]
	return value
}

func (e *basictableResource) LoadDbData(data map[string]interface{}) BasictableResourceInterface {
	for key := range e.Model.GetTableFields() {
		key = strings.ToLower(key)
		e.SetOriginData(key, nil).SetData(key, nil)
	}
	for key, value := range data {
		key = strings.ToLower(key)
		e.SetOriginData(key, value).SetData(key, value)
	}
	return e
}

func (e *basictableResource) Reset() BasictableResourceInterface {
	e.Data = make(map[string]interface{})
	e.OriginData = make(map[string]interface{})
	return e
}

func (e *basictableResource) GetDbData(isExcludeEav bool) map[string]interface{} {
	data := make(map[string]interface{})
	for key, Field := range e.Model.GetTableFields() {
		key = strings.ToLower(key)
		value, ok := e.Data[key]
		if Field.IsEav && isExcludeEav {
			continue
		}
		if ok {
			data[key] = value
		} else if key != e.Model.GetPrimaryFieldName() {
			data[key] = nil
		}
	}
	return data
}

func (e *basictableResource) GetDbOriginData() map[string]interface{} {
	data := make(map[string]interface{})
	for key := range e.Model.GetTableFields() {
		key = strings.ToLower(key)
		value, err := e.OriginData[key]
		if !err {
			data[key] = value
		} else {
			data[key] = nil
		}
	}
	return data
}
func (e *basictableResource) HasDataChange() bool {
	dbData := e.GetDbData(false)
	dbOriginData := e.GetDbOriginData()
	return !reflect.DeepEqual(dbData, dbOriginData)
}
func (e *basictableResource) autoTime() map[string]interface{} {
	data := make(map[string]interface{})
	for key, field := range e.Model.GetTableFields() {
		key = strings.ToLower(key)
		isGenTime := false
		value := e.GetData(key)
		if str, ok := value.(string); ok && (str == "" || value == nil) {
			isGenTime = true
		}
		if value == nil {
			isGenTime = true
		}
		if isGenTime && field.Autocreate && field.DbType == "time.Time" {
			e.SetData(key, time.Now().UTC())
		}
		if field.Autoupdate && field.DbType == "time.Time" {
			e.SetData(key, time.Now().UTC())
		}
	}
	return data
}

func (e *basictableResource) getEavTableByField(eavtype string) string {
	return e.Model.GetModel().GetTableName() + "_" + eavtype
}

func (e *basictableResource) saveEavFields() BasictableResourceInterface {
	locale := e.Model.GetLocale()
	if locale == "" {
		return e
	}
	eavFields := e.Model.GetEavFields()
	data := e.GetDbData(false)
	for key, field := range eavFields {
		value, ok := data[key]
		if ok {
			table := e.getEavTableByField(field.EavType)
			eavData := []map[string]interface{}{{"value": value, "entity_id": e.GetData(e.Model.GetPrimaryFieldName()), "locale": locale, "attribute_name": key}}
			e.Connection.InsertMultiOnUpdate(table, eavData)
		}
	}
	return e
}

func (e *basictableResource) Save() BasictableResourceInterface {
	if e.HasDataChange() {
		table := e.Model.GetTableName()
		e.autoTime()
		data := e.GetDbData(true)
		if e.Model.GetPrimaryFieldName() == "" {
			datas := make([]map[string]interface{}, 1)
			datas[0] = data
			e.Connection.InsertMultiOnUpdate(table, datas)
		} else {
			primaryValue := e.GetData(e.Model.GetPrimaryFieldName())
			if len(data) > 0 {
				if primaryValue == nil {
					id := e.Connection.Insert(table, data)
					e.SetData(e.Model.GetPrimaryFieldName(), id)
					e.SetOriginData(e.Model.GetPrimaryFieldName(), id)
				} else {
					// update
					sql := e.Connection.Expr("select "+e.Model.GetPrimaryFieldName()+" from "+e.Model.GetTableName()+" where "+e.Model.GetPrimaryFieldName()+"=?", primaryValue)
					dbId := e.Connection.FetchOne(sql)
					if dbId != nil {
						e.Connection.Update(table, data, e.Connection.Expr(e.Model.GetPrimaryFieldName()+"=?", primaryValue))
					} else {
						id := e.Connection.Insert(table, data)
						e.SetData(e.Model.GetPrimaryFieldName(), id)
						e.SetOriginData(e.Model.GetPrimaryFieldName(), id)
					}
				}
				e.saveEavFields()
			}
		}

	}

	return e
}

func (e *basictableResource) LoadByField(field string, value interface{}) BasictableResourceInterface {
	eavFields := e.Model.GetEavFields()
	sql := ""
	if len(eavFields) == 0 {
		sql = e.Connection.Expr("select * from "+e.Model.GetTableName()+" where "+field+"=?", value)
	} else {
		//sql := fmt.Sprintf("select m.* from %s as m ", e.Model.GetTableName())
		sql = e.Connection.Expr("select * from "+e.GetEavAsTable()+"as t where "+field+"=?", value)
	}

	data := e.Connection.FetchRow(sql)
	e.LoadDbData(data)
	return e
}

func (e *basictableResource) Delete() BasictableResourceInterface {
	table := e.Model.GetTableName()
	primaryValue := e.GetData(e.Model.GetPrimaryFieldName())
	if e.Model.GetPrimaryFieldName() != "" {
		sql := e.Connection.Expr("select "+e.Model.GetPrimaryFieldName()+" from "+e.Model.GetTableName()+" where "+e.Model.GetPrimaryFieldName()+"=?", primaryValue)
		dbId := e.Connection.FetchOne(sql)
		if dbId != nil {
			e.Connection.Delete(table, e.Connection.Expr(e.Model.GetPrimaryFieldName()+"=?", primaryValue))
		}
	} else if fields := e.Model.GetDeleteFields(); len(fields) > 0 {
		sql := ""
		values := make([]interface{}, 0)
		for _, code := range fields {
			if sql != "" {
				sql += " and "
			}
			sql += code + " =? "
			values = append(values, e.GetData(code))
		}
		e.Connection.Delete(table, e.Connection.Expr(sql, values))
	}

	return e
}

func (e *basictableResource) GetConnection() DBConnectionInterface {
	return e.Connection
}
func (e *basictableResource) GetEavAsTable() string {
	eavFields := e.Model.GetEavFields()
	sql := ""
	locale := e.Model.GetLocale()
	defalutLocale := e.Model.GetDefaultLocale()
	if eavModel, ok := e.Model.GetModel().(EavModelInterface); ok {
		return eavModel.GetEavAsTable(locale, defalutLocale)
	}
	columns := make([]string, 0)
	for key, field := range eavFields {
		sql += fmt.Sprintf(`
			left join %s as e_%s_default on e_%s_default.entity_id = m.%s and e_%s_default.locale="%s" and e_%s_default.attribute_name = "%s"
			`, e.getEavTableByField(field.EavType), key, key, e.Model.GetPrimaryFieldName(), key, defalutLocale, key, key) +
			fmt.Sprintf(`
			left join %s as e_%s on e_%s.entity_id = m.%s and e_%s.locale="%s" and e_%s.attribute_name = "%s"
			`, e.getEavTableByField(field.EavType), key, key, e.Model.GetPrimaryFieldName(), key, locale, key, key)
		columns = append(columns, fmt.Sprintf("ifNUll(e_%s.value,e_%s_default.value) as %s", key, key, key))
	}
	sql = fmt.Sprintf("(select m.*,%s from %s as m %s )", strings.Join(columns, ","), e.Model.GetTableName(), sql)
	return sql
}
