package main

import "github.com/moremorefun/mcommon"

func main() {
	query, argMap, err := mcommon.QueryInsert().
		Ignore().
		Into("t_user").
		Columns("user_name", "city").
		Values("hao", "bj").
		Values("hao1", "bj1").
		Duplicates(
			mcommon.QueryEqRaw{
				K: "count",
				V: "count+1",
			},
			mcommon.QueryDuplicateValue("time"),
			mcommon.QueryEq{
				K: "set",
				V: 1,
			},
		).
		ToSQL()
	if err != nil {
		mcommon.Log.Fatalf("err: %s", err.Error())
	}
	mcommon.Log.Debugf("query: %s\nargMap: %#v", query, argMap)
}
