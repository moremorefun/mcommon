package mcommon

// 数据库数据类型
const (
	MySqlGoTypeString  = 1
	MySqlGoTypeInt64   = 2
	MySqlGoTypeBytes   = 3
	MySqlGoTypeFloat64 = 4
	MySqlGoTypeTime    = 5
)

// MysqlTypeToGoMap 类型转换关系
var MysqlTypeToGoMap = map[string]int64{
	"BIT":        1,
	"TEXT":       1,
	"BLOB":       3,
	"DATETIME":   5,
	"DOUBLE":     4,
	"ENUM":       1,
	"FLOAT":      4,
	"GEOMETRY":   1,
	"MEDIUMINT":  2,
	"JSON":       1,
	"INT":        2,
	"LONGTEXT":   1,
	"LONGBLOB":   3,
	"BIGINT":     2,
	"MEDIUMTEXT": 1,
	"MEDIUMBLOB": 3,
	"DATE":       5,
	"DECIMAL":    1,
	"SET":        1,
	"SMALLINT":   2,
	"BINARY":     3,
	"CHAR":       1,
	"TIME":       5,
	"TIMESTAMP":  5,
	"TINYINT":    2,
	"TINYTEXT":   1,
	"TINYBLOB":   3,
	"VARBINARY":  3,
	"VARCHAR":    1,
	"YEAR":       2,
}
