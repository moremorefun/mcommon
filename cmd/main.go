package main

import "github.com/moremorefun/mcommon"

func main() {
	query, argMap, err := mcommon.
		QueryUpdate("t_user").
		Update(
			mcommon.QueryEq{
				K: "user_name",
				V: "hao",
			},
			mcommon.QueryEqRaw{
				K: "user_cat",
				V: "hello",
			},
		).
		Where(
			mcommon.QueryEq{
				K: "user_id",
				V: "hao",
			},
		).
		ToSQL()
	if err != nil {
		mcommon.Log.Fatalf("err: %s", err.Error())
	}
	mcommon.Log.Debugf("query: %s\nargMap: %#v", query, argMap)
}
