package core

import (
	"strconv"
	"strings"
)

type DBSelectInterface interface {
	From(string, string, map[string]string) DBSelectInterface
	InnerJoin(tableAlias string, table string, condition string, columns map[string]string) DBSelectInterface
	LeftJoin(tableAlias string, table string, condition string, columns map[string]string) DBSelectInterface
	Where(string) DBSelectInterface
	Limit(int) DBSelectInterface
	Offset(int) DBSelectInterface
	Assemble() (string, error)
	Init() DBSelectInterface
	Reset() DBSelectInterface
	Order(string) DBSelectInterface
}

type DBSelectError struct {
	Message string
}

func (e *DBSelectError) Error() string {
	return e.Message
}

type DBSelect struct {
	_from   map[string]interface{}
	_join   []map[string]interface{}
	_where  []string
	_limit  int
	_offset int
	_order  []string
}

func (this *DBSelect) From(table string, tableAlias string, columns map[string]string) DBSelectInterface {
	from := make(map[string]interface{})
	from["table"] = table
	from["tableAlias"] = tableAlias
	from["columns"] = columns
	this._from = from
	return this
}
func (this *DBSelect) hasJoin(tableAlias string) bool {
	for _, join := range this._join {
		if join["tableAlias"] == tableAlias {
			return true
		}
	}
	return false
}

func (this *DBSelect) InnerJoin(tableAlias string, table string, condition string, columns map[string]string) DBSelectInterface {
	if this.hasJoin(tableAlias) {
		return this
	}
	join := make(map[string]interface{})
	join["tableAlias"] = tableAlias
	join["table"] = table
	join["condition"] = condition
	join["columns"] = columns
	join["join_type"] = "inner"
	this._join = append(this._join, join)
	return this
}
func (this *DBSelect) LeftJoin(tableAlias string, table string, condition string, columns map[string]string) DBSelectInterface {
	if this.hasJoin(tableAlias) {
		return this
	}
	join := make(map[string]interface{})
	join["tableAlias"] = tableAlias
	join["table"] = table
	join["condition"] = condition
	join["columns"] = columns
	join["join_type"] = "left"
	this._join = append(this._join, join)
	return this
}
func (this *DBSelect) Where(condition string) DBSelectInterface {
	this._where = append(this._where, condition)
	return this
}
func (this *DBSelect) Limit(limit int) DBSelectInterface {
	this._limit = limit
	return this
}
func (this *DBSelect) Offset(offset int) DBSelectInterface {
	this._offset = offset
	return this
}
func (this *DBSelect) Assemble() (string, error) {
	// 處理from
	if this._from == nil {
		return "", &DBSelectError{"from is null"}
	}

	tableAlias, ok := this._from["tableAlias"].(string)
	table, ok1 := this._from["table"].(string)
	if ok1 && table == "" {
		return "", &DBSelectError{"from table  is null"}
	}

	selectStr := "select "

	if this._from["columns"] == nil {
		if tableAlias != "" {
			selectStr += tableAlias + ".* "
		} else {
			selectStr += table + ".* "
		}
	} else {
		columns, ok := this._from["columns"].(map[string]string)
		if ok && len(columns) > 0 {
			for alias, field := range columns {
				if alias == field {
					if tableAlias != "" {
						selectStr += tableAlias + field + ", "
					} else {
						selectStr += table + field + ", "
					}
				} else {
					if tableAlias != "" {
						selectStr += tableAlias + field + " as " + alias + ", "
					} else {
						selectStr += table + field + " as " + alias + ", "
					}
				}
			}
			selectStr = strings.TrimSuffix(selectStr, ", ")
		} else {
			if tableAlias != "" {
				selectStr += tableAlias + ".* "
			} else {
				selectStr += table + ".* "
			}
		}
	}

	// 處理join的cloumns
	if len(this._join) > 0 {
		for _, join := range this._join {
			joinTableAlias, _ := join["tableAlias"].(string)
			joinTable, _ := join["table"].(string)
			columns, _ := join["columns"].(map[string]string)
			if len(columns) > 0 {
				selectStr += ", "
				for alias, field := range columns {
					if joinTableAlias != "" {
						selectStr += joinTableAlias + "." + field
					} else {
						selectStr += joinTable + "." + field
					}
					if alias == field {
						selectStr += " as " + alias
					}
					selectStr += ", "
				}
				selectStr = strings.TrimSuffix(selectStr, ", ")
			}
		}
	}
	// 處理from table
	if ok1 && table == "" {
		return "", &DBSelectError{"from table  is null"}
	}
	selectStr += " from " + table
	if ok && tableAlias != "" {
		selectStr += " as " + tableAlias + " "
	} else {
		selectStr += " "
	}

	// 處理join table
	if len(this._join) > 0 {
		for _, join := range this._join {
			joinTableAlias, _ := join["tableAlias"].(string)
			joinTable, _ := join["table"].(string)
			condition, _ := join["condition"].(string)
			joinType, _ := join["join_type"].(string)
			switch joinType {
			case "inner":
				selectStr += " inner join "
				break
			case "left":
				selectStr += " left join "
				break
			}
			if joinTableAlias != "" {
				selectStr += joinTable + " as  " + joinTableAlias + " on " + condition
			} else {
				selectStr += joinTable + " on " + condition
			}
		}
	}

	// 處理where
	if len(this._where) > 0 {
		selectStr += " where "
		for _, where := range this._where {
			selectStr += where + " and "
		}
		selectStr = strings.TrimSuffix(selectStr, "and ")
	}
	// 處理order by
	if len(this._order) > 0 {
		selectStr += " order by "
		for _, order := range this._order {
			selectStr += order + ", "
		}
		selectStr = strings.TrimSuffix(selectStr, ", ")
	}

	//處理limit
	if this._limit > 0 {
		selectStr += " limit " + strconv.Itoa(this._limit)
	}

	// 處理offset
	if this._offset > 0 {
		selectStr += " offset " + strconv.Itoa(this._offset)
	}

	return selectStr, nil
}
func (this *DBSelect) Init() DBSelectInterface {
	this.Reset()
	return this
}
func (this *DBSelect) Reset() DBSelectInterface {
	this._from = make(map[string]interface{})
	this._join = make([]map[string]interface{}, 0)
	this._limit = 0
	this._offset = 0
	this._order = make([]string, 0)

	return this
}
func (this *DBSelect) Order(order string) DBSelectInterface {
	this._order = append(this._order, order)
	return this
}
