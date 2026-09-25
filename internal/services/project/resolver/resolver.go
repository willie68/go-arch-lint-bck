package resolver

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/willie68/go-arch-lint/internal/models"
	"github.com/willie68/go-arch-lint/internal/models/arch"
)

type Resolver struct {
	projectFilesResolver projectFilesResolver
	projectFilesHolder   projectFilesHolder
}

func NewResolver(
	projectFilesResolver projectFilesResolver,
	projectFilesHolder projectFilesHolder,
) *Resolver {
	return &Resolver{
		projectFilesResolver: projectFilesResolver,
		projectFilesHolder:   projectFilesHolder,
	}
}

func (r *Resolver) ProjectFiles(ctx context.Context, spec arch.Spec) ([]models.FileHold, error) {
	scanDirectory := filepath.Join(
		spec.RootDirectory.Value,
		filepath.FromSlash(spec.WorkingDirectory.Value),
	)

	projectFiles, err := r.projectFilesResolver.Scan(
		ctx,
		scanDirectory,
		spec.ModuleName.Value,
		refPathToList(spec.Exclude),
		refRegExpToList(spec.ExcludeFilesMatcher),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project files: %w", err)
	}

	holdFiles := r.projectFilesHolder.HoldProjectFiles(projectFiles, spec.Components)
	return holdFiles, nil
}
