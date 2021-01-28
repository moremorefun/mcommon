package mcommon

var MysqlTypeToGoMap = map[string]string{
	"BIT":        "string",
	"TEXT":       "string",
	"BLOB":       "byte",
	"DATETIME":   "time",
	"DOUBLE":     "float64",
	"ENUM":       "string",
	"FLOAT":      "float64",
	"GEOMETRY":   "string",
	"MEDIUMINT":  "int64",
	"JSON":       "string",
	"INT":        "int64",
	"LONGTEXT":   "string",
	"LONGBLOB":   "byte",
	"BIGINT":     "int64",
	"MEDIUMTEXT": "string",
	"MEDIUMBLOB": "byte",
	"DATE":       "time",
	"DECIMAL":    "string",
	"SET":        "string",
	"SMALLINT":   "int64",
	"BINARY":     "byte",
	"CHAR":       "string",
	"TIME":       "time",
	"TIMESTAMP":  "time",
	"TINYINT":    "int64",
	"TINYTEXT":   "string",
	"TINYBLOB":   "byte",
	"VARBINARY":  "byte",
	"VARCHAR":    "string",
	"YEAR":       "int64",
}
