package sql2struct

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var DBTypeToStructType = map[string]string{
	"int": "int32", "tinyint": "int8", "smallint": "int", "mediumint": "int64", "bigint": "int64", "bit": "int", "bool": "bool", "enum": "string", "set": "string", "varchar": "string",
}

type DbModel struct {
	DBEngine *sql.DB
	DBInfo   *DBInfo
}

type DBInfo struct {
	DBType   string
	Host     string
	Username string
	Password string
	Charset  string
}

type TableColumn struct {
	ColumnName    string
	DataType      string
	IsNullAble    string
	ColumnKey     string
	ColumnType    string
	ColumnComment string
}

func NewDBModel(info *DBInfo) *DbModel {
	return &DbModel{
		DBInfo: info,
	}
}

func (m *DbModel) Connect() error {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/information_schema?charset=%s&parseTime=True&loc=Local", m.DBInfo.Username, m.DBInfo.Password, m.DBInfo.Host, m.DBInfo.Charset)
	m.DBEngine, err = sql.Open(m.DBInfo.DBType, dsn)
	if err != nil {
		return err
	}
	return nil
}

func (m *DbModel) GetColumns(dbName, tableName string) ([]*TableColumn, error) {
	query := "SELECT COLUMN_NAME, DATA_TYPE, COLUMN_KEY, " +
		"IS_NULLABLE, COLUMN_TYPE, COLUMN_COMMENT " +
		"FROM COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? "
	rows, err := m.DBEngine.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, errors.New("数据空")
	}
	defer rows.Close()
	var columns []*TableColumn
	for rows.Next() {
		var column TableColumn
		err := rows.Scan(&column.ColumnName, &column.ColumnType, &column.ColumnKey, &column.ColumnComment, &column.DataType, &column.IsNullAble)
		if err != nil {
			return nil, err
		}
		columns = append(columns, &column)
	}
	return columns, nil
}
