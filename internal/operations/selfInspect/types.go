package selfInspect

import (
	"github.com/willie68/go-arch-lint/internal/models/arch"
	"github.com/willie68/go-arch-lint/internal/models/common"
)

type (
	specAssembler interface {
		Assemble(prj common.Project) (arch.Spec, error)
	}

	projectInfoAssembler interface {
		ProjectInfo(rootDirectory string, archFilePath string) (common.Project, error)
	}
)
