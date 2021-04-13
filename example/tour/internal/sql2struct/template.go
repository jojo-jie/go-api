package sql2struct

import (
	"fmt"
	"os"
	"text/template"
	"tour/internal/word"
)

const structTpl = `type {{.TableName | ToCamelCase}} struct {
{{range .Columns}} {{ $length := len .Comment}} {{ if gt $len
gth 0 }}//{{.Comment}} {{else}}// {{.Name}} {{ end }}    {{ $typeLen := len .Type }} {{ if gt $typeLen 0 }}{{.Name
 | ToCamelCase}}    {{.Type}}     {{.Tag}}{{ else }}{{.Name}}{{ end }}{{end}}}func (model {{.TableName | ToCamelCase}}) TableName() string 
{    return "{{.TableName}}"}`

type StructTemplate struct {
	structTpl string
}

type StructColumn struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

type StructTemplateDB struct {
	TableName string
	Columns   []*StructColumn
}

func NewStructTemplate() *StructTemplate {
	return &StructTemplate{structTpl: structTpl}
}

// AssemblyColumns 渲染模板
func (t *StructTemplate) AssemblyColumns(tbColumns []*TableColumn) []*StructColumn {
	tplColumns := make([]*StructColumn, 0, len(tbColumns))
	for _, column := range tbColumns {
		tag := fmt.Sprintf("`"+"json:"+"\"%s\""+"`", column.ColumnName)
		tplColumns = append(tplColumns, &StructColumn{
			Name:    column.ColumnName,
			Type:    DBTypeToStructType[column.ColumnType],
			Tag:     tag,
			Comment: column.ColumnComment,
		})
	}
	return tplColumns
}

func (t *StructTemplate) Generate(tableName string, tplColumns []*StructColumn) error {
	tpl := template.Must(template.New("sql2struct").Funcs(template.FuncMap{
		"ToCamelCase": word.UnderscoreToUpperCamelCase,
	}).Parse(t.structTpl))
	tplDB := StructTemplateDB{
		tableName, tplColumns,
	}
	return tpl.Execute(os.Stdout, tplDB)
}
