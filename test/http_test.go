package test

import (
	"fmt"
	"strings"
	"testing"
)

func TestStringSplit(t *testing.T) {
	basePath := "/_zcache/"
	path := "/_zcache/test_group/key1"
	parts := strings.SplitN(path[len(basePath):], "/", 2)
	fmt.Println(parts)

	s := "1/2/3/4/5/6/7"
	ret := strings.SplitN(s, "/", 3)
	fmt.Println(ret)
}
