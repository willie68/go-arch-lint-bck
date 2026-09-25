package a

import "github.com/willie68/go-arch-lint/test/check/project/internal/common"

func A1() {
	common.C1() // common - allowed
}
