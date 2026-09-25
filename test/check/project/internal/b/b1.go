package b

import "github.com/willie68/go-arch-lint/test/check/project/internal/common/sub/foo/bar"

func B1() {
	bar.BR1() // common - allowed
}
