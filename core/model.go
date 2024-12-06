package core

type BasicModelInterface interface {
	GetTableName() string
	GetTableFields() map[string]Field
	GetPrimaryFieldName() string
}
type EavModelInterface interface {
	GetEavAsTable(locale string, defaultLocale string) string
}
type BasicModelLoadInterface interface {
	AfterLoad(Basictablemodelinterface)
}
type BasicModelDeleteInterface interface {
	AfterDelete(Basictablemodelinterface)
}
type BasicModelSaveInterface interface {
	AfterSave(Basictablemodelinterface)
}
type CollectionFieldInterface interface {
	AddJoinField(collection CollectionInterface, field string) string
}

func ModelFactory(callback func() Basictablemodelinterface) Basictablemodelinterface {
	tableModel := callback()
	tableModel.Init()
	return tableModel
}
