package main

import (
	"fmt"
	"testing"
)

func TestOne(t *testing.T) {
	err := updateDnsRecord(MyInfo{

	},
		"6.5.6.6")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("成功")
	}
}
