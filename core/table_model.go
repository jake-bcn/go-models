package core

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Field struct {
	Name       string
	IsEav      bool
	DbType     string
	Autocreate bool
	Autoupdate bool
	EavType    string
}
type Basictablemodelinterface interface {
	Save() Basictablemodelinterface
	Delete() Basictablemodelinterface
	LoadByField(string, interface{}) Basictablemodelinterface
	LoadById(id interface{}) Basictablemodelinterface
	GetTableName() string
	GetTableFields() map[string]Field
	GetPrimaryFieldName() string
	Init() Basictablemodelinterface
	GetResourceModel() BasictableResourceInterface
	SetData(string, interface{}) Basictablemodelinterface
	GetData(string) interface{}
	GetModel() BasicModelInterface
	GetConnection() DBConnectionInterface
	GetConnectionName() string
	GetLocale() string
	GetEavFields() map[string]Field
	GetDefaultLocale() string
	GetLastError() error
	GetDeleteFields() []string
}
type Basictablemodel struct {
	ResourceModel *basictableResource
	Model         BasicModelInterface
	Connection    string
	Locale        string
	DefaultLocale string
	LastError     error
}

func (e *Basictablemodel) GetLastError() error {
	return e.LastError
}

func (e *Basictablemodel) _transaction(callback func()) {
	db := GetConnection(e.GetConnectionName())
	e.LastError = nil
	if db != nil {
		db.Transaction(func(tx *gorm.DB) error {
			defer func() {
				if r := recover(); r != nil {
					switch x := r.(type) {
					case string:
						e.LastError = errors.New(x)
					case error:
						e.LastError = x
					default:
						e.LastError = fmt.Errorf("unexpected error: %v", r)
					}
				}
			}()
			callback()
			return nil
		})
	}
}

func (e *Basictablemodel) Save() Basictablemodelinterface {
	e._transaction(func() {
		if m, ok := interface{}(e.Model).(BasicModelBeforeSaveInterface); ok {
			m.BeforeSave(e)
		}
		e.ResourceModel.Save()
		if m, ok := interface{}(e.Model).(BasicModelSaveInterface); ok {
			m.AfterSave(e)
		}
	})
	return e
}

func (e *Basictablemodel) Delete() Basictablemodelinterface {
	e._transaction(func() {
		if m, ok := interface{}(e.Model).(BasicModelDeleteBeforeInterface); ok {
			m.BeforeDelete(e)
		}
		e.ResourceModel.Delete()
		if m, ok := interface{}(e.Model).(BasicModelDeleteInterface); ok {
			m.AfterDelete(e)
		}
	})

	return e
}
func (e *Basictablemodel) LoadByField(field string, value interface{}) Basictablemodelinterface {
	e._transaction(func() {
		e.ResourceModel.LoadByField(field, value)
		if m, ok := interface{}(e.Model).(BasicModelLoadInterface); ok {
			m.AfterLoad(e)
		}
	})

	return e
}
func (e *Basictablemodel) LoadById(id interface{}) Basictablemodelinterface {
	e.LoadByField(e.GetPrimaryFieldName(), id)
	return e
}

func (e *Basictablemodel) GetTableName() string {
	return e.Model.GetTableName()
}

func (e *Basictablemodel) GetTableFields() map[string]Field {
	return e.Model.GetTableFields()
}
func (e *Basictablemodel) Init() Basictablemodelinterface {
	e.ResourceModel = &basictableResource{}
	e.ResourceModel.Initialize(e, e.GetConnectionName())
	return e
}
func (e *Basictablemodel) GetResourceModel() BasictableResourceInterface {
	return e.ResourceModel
}

func (e *Basictablemodel) SetData(field string, value interface{}) Basictablemodelinterface {
	e.GetResourceModel().SetData(field, value)
	return e
}
func (e *Basictablemodel) GetModel() BasicModelInterface {
	return e.Model
}

func (e *Basictablemodel) GetData(field string) interface{} {
	value := e.GetResourceModel().GetData(field)
	return value
}

func (e *Basictablemodel) GetPrimaryFieldName() string {
	return strings.ToLower(e.Model.GetPrimaryFieldName())
}
func (e *Basictablemodel) GetConnection() DBConnectionInterface {
	return e.GetResourceModel().GetConnection()
}
func (e *Basictablemodel) GetConnectionName() string {
	if e.Connection != "" {
		return e.Connection
	}
	return "default"
}
func (e *Basictablemodel) GetLocale() string {
	return e.Locale
}
func (e *Basictablemodel) GetDefaultLocale() string {
	return e.DefaultLocale
}
func (e *Basictablemodel) GetEavFields() map[string]Field {
	fields := make(map[string]Field)
	for key, field := range e.Model.GetTableFields() {
		if field.IsEav && field.EavType != "" {
			fields[key] = field
		}
	}
	return fields
}
func (e *Basictablemodel) GetDeleteFields() []string {
	var result []string
	if m, ok := interface{}(e.Model).(BasicModelGetDeleteFieldsInterface); ok {
		return m.GetDeleteFields(e)
	}
	return result
}
