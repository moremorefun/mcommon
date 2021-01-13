package mcommon

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	QueryJoinTypeInner = 1
)

type QueryMaker interface {
	ToSQL() ([]byte, map[string]interface{}, error)
}

func getK(old string) string {
	old = strings.ReplaceAll(old, ".", "_")
	old = strings.ReplaceAll(old, "`", "_")
	return old
}

type QueryKv struct {
	K string
	V interface{}
}

type QueryKvStr struct {
	K string
	V string
}

type QueryEq QueryKv

func (o QueryEq) ToSQL() ([]byte, map[string]interface{}, error) {
	k := getK(o.K)

	var buf bytes.Buffer
	args := map[string]interface{}{}
	buf.WriteString(o.K)
	rt := reflect.TypeOf(o.V)
	switch rt.Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(o.V)
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
	args[k] = o.V
	return buf.Bytes(), args, nil
}

type QueryGt QueryKv

func (o QueryGt) ToSQL() ([]byte, map[string]interface{}, error) {
	k := getK(o.K)

	var buf bytes.Buffer
	args := map[string]interface{}{}
	buf.WriteString(o.K)
	buf.WriteString(">:")
	buf.WriteString(k)
	args[k] = o.V
	return buf.Bytes(), args, nil
}

type QueryLt QueryKv

func (o QueryLt) ToSQL() ([]byte, map[string]interface{}, error) {
	k := getK(o.K)

	var buf bytes.Buffer
	args := map[string]interface{}{}
	buf.WriteString(o.K)
	buf.WriteString("<:")
	buf.WriteString(k)
	args[k] = o.V
	return buf.Bytes(), args, nil
}

type QueryDesc string

func (o QueryDesc) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	buf.WriteString(string(o))
	buf.WriteString(" DESC")
	return buf.Bytes(), nil, nil
}

type QueryAsc string

func (o QueryAsc) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	buf.WriteString(string(o))
	return buf.Bytes(), nil, nil
}

type QueryEqColumn QueryKvStr

func (o QueryEqColumn) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	buf.WriteString(o.K)
	buf.WriteString("=")
	buf.WriteString(o.V)
	return buf.Bytes(), nil, nil
}

type joinData struct {
	joinType int64
	obj      string
	onParts  []QueryMaker
}

// QueryJoin 链接
func QueryJoin(joinType int64, obj string) *joinData {
	j := joinData{
		joinType: joinType,
		obj:      obj,
	}
	return &j
}

// On 链接条件
func (j *joinData) On(cond QueryMaker) *joinData {
	j.onParts = append(j.onParts, cond)
	return j
}

// ToSQL 生成sql
func (j *joinData) ToSQL() ([]byte, map[string]interface{}, error) {
	var buf bytes.Buffer
	args := map[string]interface{}{}

	switch j.joinType {
	case QueryJoinTypeInner:
		buf.WriteString("INNER JOIN")
	default:
		return nil, nil, fmt.Errorf("no joinData type: %d", j.joinType)
	}
	if len(j.obj) == 0 {
		return nil, nil, fmt.Errorf("joinData obj emputy")
	}
	buf.WriteString(" ")
	buf.WriteString(j.obj)
	buf.WriteString(" ON (")
	if len(j.onParts) == 0 {
		return nil, nil, fmt.Errorf("no joinData on condiation")
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
	whereParts   []QueryMaker
	groupBys     []string
	orderByParts []QueryMaker
	offset       int64
	limit        int64
	isForUpdate  bool
	joins        []QueryMaker
}

// QuerySelect 创建搜索
func QuerySelect(columns ...string) *selectData {
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
func (q *selectData) Where(cond QueryMaker) *selectData {
	q.whereParts = append(q.whereParts, cond)
	return q
}

// GroupBy 分组
func (q *selectData) GroupBy(groupBys ...string) *selectData {
	q.groupBys = groupBys
	return q
}

// OrderBy 排序
func (q *selectData) OrderBy(order ...QueryMaker) *selectData {
	q.orderByParts = append(q.orderByParts, order...)
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

// QueryJoin 链接
func (q *selectData) Join(join QueryMaker) *selectData {
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
	if len(q.joins) > 0 {
		for _, join := range q.joins {
			tQuery, tArgMap, err := join.ToSQL()
			if err != nil {
				return nil, nil, err
			}
			buf.WriteString("\n")
			buf.Write(tQuery)
			for tK, tV := range tArgMap {
				args[tK] = tV
			}
		}
	}
	if len(q.whereParts) > 0 {
		buf.WriteString("\nWHERE")
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
			buf.WriteString(strconv.FormatInt(q.offset, 10))
			buf.WriteString(", ")
			buf.WriteString(strconv.FormatInt(q.limit, 10))
		} else {
			buf.WriteString("\nLIMIT ")
			buf.WriteString(strconv.FormatInt(q.limit, 10))
		}
	}
	if q.isForUpdate {
		buf.WriteString("\nFOR UPDATE")
	}
	return buf.Bytes(), args, nil
}
