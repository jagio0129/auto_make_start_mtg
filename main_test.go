package main

import (
	"testing"
	"fmt"
	"io/ioutil"
)

func Test_mkTxt(t *testing.T) {
	var membList []string
	membList = append(membList, "山田太郎")
	membList = append(membList, "山田 太郎") // 半角スペース
	membList = append(membList, "山田　太郎") // 全角スペース
	membList = append(membList, "山田太郎[削除]")
	txt := mkTxt(membList)

	d, err := ioutil.ReadFile("./test.md")
	if err != nil {
			t.Fatal("file read error")
	}
	expectTxt := string(d)

	if txt != expectTxt {
		fmt.Println("expect :\n", expectTxt)
		fmt.Println("txt :\n", txt)

		t.Fatal("failed test")
	}

}
