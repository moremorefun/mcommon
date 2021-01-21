package main

import "github.com/moremorefun/mcommon"

func main() {
	resp, err := mcommon.Kuaidi100Poll(
		"123",
		"shunfeng",
		"SF1324271850710",
		"15210004756",
		"https://8ee552b3b302b8e11717e980be7d6d61.m.pipedream.net",
	)
	if err != nil {
		mcommon.Log.Fatalf("err: %s", err.Error())
	}
	mcommon.Log.Debugf("%#v", resp)
}
