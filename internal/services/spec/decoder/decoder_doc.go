package decoder

import "github.com/willie68/go-arch-lint/internal/services/spec"

type doc interface {
	spec.Document

	postSetup()
}
