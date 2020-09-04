package mcommon

import (
	"bytes"
	"context"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/schemalex/schemalex/diff"
)

// DbStructGetDiff 获取数据库更新指令
func DbStructGetDiff(tx DbExeAble, tableNames []string, sqlFilePath string) (string, error) {
	var dbSQLs []string
	for _, tableName := range tableNames {
		var row struct {
			TableName string `db:"Table"`
			TableSQL  string `db:"Create Table"`
		}
		ok, err := DbGetNamedContent(
			context.Background(),
			tx,
			&row,
			`SHOW CREATE TABLE `+tableName,
			gin.H{},
		)
		if err != nil {
			if strings.Contains(err.Error(), "doesn't exist") {
				continue
			}
			return "", err
		}
		if ok {
			dbSQLs = append(dbSQLs, row.TableSQL+";")
		}
	}
	// 原始sql
	dbSQL := strings.Join(dbSQLs, "\n")
	// 目的sql
	toSQL, err := ioutil.ReadFile(sqlFilePath)
	if err != nil {
		return "", err
	}
	sqlDiff := new(bytes.Buffer)
	err = diff.Strings(sqlDiff, dbSQL, string(toSQL), diff.WithTransaction(true))
	if err != nil {
		return "", err
	}
	// 替换 AUTO_INCREMENT
	r, _ := regexp.Compile(`AUTO_INCREMENT\s*=\s*(\d)*\s*,`)
	sqlDiffWithoutInc := r.ReplaceAllStringFunc(sqlDiff.String(), func(s string) string {
		return ""
	})
	return sqlDiffWithoutInc, nil
}
