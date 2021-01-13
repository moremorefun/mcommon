package mcommon

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	JoinTypeInner = 1
)

func getK(old string) string {
	old = strings.ReplaceAll(old, ".", "_")
	old = strings.ReplaceAll(old, "`", "_")
	return old
}

type kv struct {
	k string
	v interface{}
}

type kvStr struct {
	k string
	v string
}

type SQLMaker interface {
	ToSQL() ([]byte, map[string]interface{}, error)
}

type Eq kv

func (o Eq) ToSQL() ([]byte, map[string]interface{}, error) {
	k := getK(o.k)

	var buf bytes.Buffer
	args := map[string]interface{}{}
	buf.WriteString(o.k)
	rt := reflect.TypeOf(o.v)
	switch rt.Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(o.v)
		if s.Len() == 0 {
			return nil, nil, fmt.Errorf("in cond len 0")
		}
		buf.WriteString(" IN (:")
		buf.WriteString(k)
		buf.WriteString(")")
	default:
		buf.WriteString("=:")
		buf.WriteString(k)
	}
	args[k] = o.v
	return buf.Bytes(), args, nil
}

type Desc string

func (o Desc) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	buf.WriteString(string(o))
	buf.WriteString(" DESC")
	return buf.Bytes(), nil, nil
}

type Asc string

func (o Asc) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	buf.WriteString(string(o))
	return buf.Bytes(), nil, nil
}

type EqColumn kvStr

func (o EqColumn) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	buf.WriteString(o.k)
	buf.WriteString("=")
	buf.WriteString(o.v)
	return buf.Bytes(), nil, nil
}

type join struct {
	joinType int64
	obj      string
	onParts  []SQLMaker
}

// Join 链接
func Join(joinType int64, obj string) *join {
	j := join{
		joinType: joinType,
		obj:      obj,
	}
	return &j
}

// On 链接条件
func (j *join) On(cond SQLMaker) *join {
	j.onParts = append(j.onParts, cond)
	return j
}

// ToSQL 生成sql
func (j *join) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	args := map[string]interface{}{}

	switch j.joinType {
	case JoinTypeInner:
		buf.WriteString("INNER JOIN")
	default:
		return nil, nil, fmt.Errorf("no join type: %d", j.joinType)
	}
	if len(j.obj) == 0 {
		return nil, nil, fmt.Errorf("join obj emputy")
	}
	buf.WriteString(" ")
	buf.WriteString(j.obj)
	buf.WriteString(" ON (")
	if len(j.onParts) == 0 {
		return nil, nil, fmt.Errorf("no join on condiation")
	}
	for i, on := range j.onParts {
		buf.WriteString("\n    ")
		if i != 0 {
			buf.WriteString("AND ")
		}
		tQuery, tArgMap, err := on.ToSQL()
		if err != nil {
			return nil, nil, err
		}
		buf.Write(tQuery)
		for tk, tv := range tArgMap {
			args[tk] = tv
		}
	}
	buf.WriteString("\n)")
	return buf.Bytes(), args, nil
}

type selectData struct {
	columns      []string
	from         string
	whereParts   []SQLMaker
	groupBys     []string
	orderByParts []SQLMaker
	offset       int64
	limit        int64
	isForUpdate  bool
	joins        []SQLMaker
}

// Select 创建搜索
func Select(columns ...string) *selectData {
	var q selectData
	if len(columns) == 0 {
		q.columns = []string{"*"}
	} else {
		q.columns = columns
	}
	return &q
}

// From 表名
func (q *selectData) From(from string) *selectData {
	q.from = from
	return q
}

// Where 条件
func (q *selectData) Where(cond SQLMaker) *selectData {
	q.whereParts = append(q.whereParts, cond)
	return q
}

// GroupBy 分组
func (q *selectData) GroupBy(groupBys ...string) *selectData {
	q.groupBys = groupBys
	return q
}

// OrderBy 排序
func (q *selectData) OrderBy(order SQLMaker) *selectData {
	q.orderByParts = append(q.orderByParts, order)
	return q
}

// Limit 限制
func (q *selectData) Limit(limit int64) *selectData {
	q.limit = limit
	return q
}

// Offset 偏移
func (q *selectData) Offset(offset int64) *selectData {
	q.offset = offset
	return q
}

// Join 链接
func (q *selectData) Join(join SQLMaker) *selectData {
	q.joins = append(q.joins, join)
	return q
}

// ForUpdate 加锁
func (q *selectData) ForUpdate() *selectData {
	q.isForUpdate = true
	return q
}

// ToSQL 生成sql
func (q *selectData) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	args := map[string]interface{}{}
	buf.WriteString("SELECT")
	lastColumnIndex := len(q.columns) - 1
	for i, column := range q.columns {
		buf.WriteString("\n    ")
		buf.WriteString(column)
		if i != lastColumnIndex {
			buf.WriteString(",")
		}
	}
	if len(q.from) == 0 {
		return nil, nil, fmt.Errorf("select no from")
	}
	buf.WriteString("\nFROM\n    ")
	buf.WriteString(q.from)
	if len(q.whereParts) > 0 {
		buf.WriteString("WHERE")
	}
	for i, where := range q.whereParts {
		buf.WriteString("\n    ")
		if i != 0 {
			buf.WriteString("AND ")
		}
		tQuery, tArgMap, err := where.ToSQL()
		if err != nil {
			return nil, nil, err
		}
		buf.Write(tQuery)
		for tk, tv := range tArgMap {
			args[tk] = tv
		}
	}
	if len(q.groupBys) > 0 {
		buf.WriteString("\nGROUP BY\n    ")
		for i, groupBy := range q.groupBys {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(groupBy)
		}
	}
	if len(q.orderByParts) > 0 {
		buf.WriteString("\nORDER BY\n    ")
		for i, orderByPart := range q.orderByParts {
			if i != 0 {
				buf.WriteString(", ")
			}
			tQuery, _, err := orderByPart.ToSQL()
			if err != nil {
				return nil, nil, err
			}
			buf.Write(tQuery)
		}
	}
	if q.limit > 0 {
		if q.offset > 0 {
			buf.WriteString("\nLIMIT ")
			buf.WriteString(strconv.FormatInt(q.offset, 0))
			buf.WriteString(", ")
			buf.WriteString(strconv.FormatInt(q.limit, 0))
		} else {
			buf.WriteString("\nLIMIT ")
			buf.WriteString(strconv.FormatInt(q.limit, 0))
		}
	}
	if q.isForUpdate {
		buf.WriteString("\nFOR UPDATE")
	}
	return buf.Bytes(), args, nil
}
