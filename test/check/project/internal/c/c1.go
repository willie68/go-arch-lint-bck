package c

import "github.com/willie68/go-arch-lint/test/check/project/internal/a"

func C1() {
	a.A1() // not allowed
}
