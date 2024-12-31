package core

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type CollectionInterface interface {
	Load() CollectionInterface
	GetSize() int
	GetConnection() DBConnectionInterface
	GetElems() []Basictablemodelinterface
	Initialize(factory func() Basictablemodelinterface) CollectionInterface // 構造model 的工廠
	SetPageSize(int) CollectionInterface
	SetPage(int) CollectionInterface
	GetPageSize() int
	GetPageLength() int
	GetPage() int
	GetSelect() DBSelectInterface
	AddFieldToFilter(map[string]map[string]interface{}) CollectionInterface // 格式： {"name"：{"="："abc"}},  {"name"：{"="："abc"},"price":{">":100}}
	AddFieldToSelect(string) CollectionInterface
	AddOrder(order string, dir string) CollectionInterface
	Create() Basictablemodelinterface
	GetLastError() error
}

type Collection struct {
	Elems              []Basictablemodelinterface
	Connection         DBConnectionInterface
	DbSelect           DBSelectInterface
	IsLoad             bool
	IsSizeLoad         bool
	Size               int
	PageSize           int
	PageLength         int
	Page               int
	Factory            func() Basictablemodelinterface
	Model              Basictablemodelinterface
	ColumnsOfMaintable []string
	LastError          error
}

func (e *Collection) GetLastError() error {
	return e.LastError
}

func (e *Collection) _transaction(callback func()) {
	db := GetConnection(e.Model.GetConnectionName())
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

func (e *Collection) Load() CollectionInterface {
	if !e.IsLoad {
		e._transaction(func() {
			columns := make(map[string]string)
			if len(e.ColumnsOfMaintable) > 0 {
				for _, value := range e.ColumnsOfMaintable {
					columns[value] = value
				}
			}
			eavFields := e.Model.GetEavFields()
			if len(eavFields) == 0 {
				e.DbSelect.From(e.Model.GetTableName(), "e", columns)
			} else {
				e.DbSelect.From(e.Model.GetResourceModel().GetEavAsTable(), "e", columns)
			}
			e.DbSelect.Offset((e.Page - 1) * e.PageSize)
			if e.PageSize > 0 {
				e.DbSelect.Limit(e.PageSize)
			}

			sql, _ := e.DbSelect.Assemble()
			rows := e.Connection.Fetch(sql)
			for _, row := range rows {
				model := e.Factory().Init()
				model.GetResourceModel().LoadDbData(row)
				if m, ok := interface{}(model.GetModel()).(BasicModelLoadInterface); ok {
					m.AfterLoad(model)
				}
				e.Elems = append(e.Elems, model)
			}
			e.IsLoad = true
		})

	}
	return e
}

func (e *Collection) GetSize() int {
	if !e.IsSizeLoad {
		e._transaction(func() {
			e.Load()
			dbselect, _ := e.DbSelect.(*DBSelect)
			clonedbselectObj := *dbselect
			clonedbselect := &clonedbselectObj
			clonedbselect.Limit(0).Offset(0)

			rawsql, _ := clonedbselect.Assemble()
			sql := "select count(*) from (" + rawsql + ") as t"
			res := (e.Connection.FetchOne(sql))
			cnt, ok := res.(string)
			if ok {
				e.IsSizeLoad = true
				size, _ := strconv.Atoi(cnt)
				if e.PageSize == 0 && size > 0 {
					e.PageSize = size
					e.PageLength = 1
				} else {
					e.PageLength = int(math.Ceil(float64(size) / float64(e.PageSize)))
				}
				e.Size = size
			}
		})

	}

	return e.Size
}
func (e *Collection) GetConnection() DBConnectionInterface {
	return e.Connection
}
func (e *Collection) GetElems() []Basictablemodelinterface {
	e.Load()
	return e.Elems
}
func (e *Collection) Initialize(factory func() Basictablemodelinterface) CollectionInterface {
	e.Factory = factory
	e.Model = e.Factory().Init()
	e.Connection = &DBConnection{}
	e.Connection.Init(e.Model.GetConnectionName())
	e.DbSelect = &DBSelect{}
	e.reset()
	return e
}
func (e *Collection) reset() CollectionInterface {
	e.DbSelect.Init()
	e.SetPageSize(0)
	e.Size = 0
	e.IsSizeLoad = false
	e.IsLoad = false
	e.Elems = make([]Basictablemodelinterface, 0)
	e.Size = 0
	e.ColumnsOfMaintable = make([]string, 0)
	e.Page = 1
	return e
}
func (e *Collection) SetPageSize(size int) CollectionInterface {
	e.PageSize = size
	return e
}
func (e *Collection) GetPageSize() int {
	e.GetSize()
	return e.PageSize
}
func (e *Collection) SetPage(page int) CollectionInterface {
	e.Page = page
	return e
}
func (e *Collection) GetPage() int {
	return e.Page
}
func (e *Collection) GetPageLength() int {
	e.GetSize()
	return e.PageLength
}
func (e *Collection) GetSelect() DBSelectInterface {
	return e.DbSelect
}

func (e *Collection) AddFieldToFilter(values map[string]map[string]interface{}) CollectionInterface {
	// 複雜的select 請用原生的
	sql := ""
	fieldValues := make([]interface{}, 0)
	if len(values) > 0 {
		sql += "("
		for field, fieldcondition := range values {
			sql += "("
			for condition, value := range fieldcondition {
				model, ok := e.Model.GetModel().(CollectionFieldInterface)
				if ok {
					field1 := model.AddJoinField(e, field)
					if field1 == "" {
						field = "e." + field
					} else {
						field = field1
					}
				}
				if strings.ToLower(condition) == "in" {
					sql += "(" + field + " " + condition + "(?)" + ") and"
				} else if strings.ToLower(condition) == "not in" {
					sql += "(" + field + " " + condition + "(?)" + ") and"
				} else {
					sql += "(" + field + " " + condition + "?" + ") and"
				}
				fieldValues = append(fieldValues, value)
			}
			sql = strings.TrimSuffix(sql, "and") + ") or"
		}
		sql = strings.TrimSuffix(sql, "or") + ")"
	}
	sql = e.Connection.Expr(sql, fieldValues...)
	e.DbSelect.Where(sql)

	return e
}

func (e *Collection) AddFieldToSelect(field string) CollectionInterface {
	_, ok := e.Model.GetTableFields()[field]
	if ok {
		e.ColumnsOfMaintable = append(e.ColumnsOfMaintable, field)
	}

	model, ok1 := e.Model.GetModel().(CollectionFieldInterface)
	if ok1 {
		model.AddJoinField(e, field)
	}
	return e
}

func (e *Collection) AddOrder(order string, dir string) CollectionInterface {
	model, ok := e.Model.GetModel().(CollectionFieldInterface)
	if ok {
		field1 := model.AddJoinField(e, order)
		if field1 == "" {
			order = "e." + order
		} else {
			order = field1
		}
	}
	e.DbSelect.Order(fmt.Sprintf("%s %s", order, dir))

	return e
}
func (e *Collection) Create() Basictablemodelinterface {
	return e.Factory()
}

func CollectionFactory(callback func() Basictablemodelinterface) CollectionInterface {
	ACollection := &Collection{}
	ACollection.Initialize(callback)
	return ACollection
}
