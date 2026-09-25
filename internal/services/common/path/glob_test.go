package path

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlob_trimsSeparatorsAroundDoubleStar(t *testing.T) {
	root := t.TempDir()

	servicesDir := filepath.Join(root, "app", "internal", "services")
	modelsDir := filepath.Join(root, "app", "internal", "models")
	require.NoError(t, os.MkdirAll(servicesDir, 0o755))
	require.NoError(t, os.MkdirAll(modelsDir, 0o755))

	foo := filepath.Join(servicesDir, "foo.go")
	bar := filepath.Join(modelsDir, "bar.go")
	require.NoError(t, os.WriteFile(foo, []byte("package services\n"), 0o600))
	require.NoError(t, os.WriteFile(bar, []byte("package models\n"), 0o600))

	internal := filepath.Join(root, "app", "internal")
	app := filepath.Join(root, "app")
	underInternal := []string{internal, servicesDir, modelsDir, foo, bar}

	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "os separator before double star",
			pattern: internal + string(filepath.Separator) + "**",
			want:    underInternal,
		},
		{
			name:    "unix slash before double star",
			pattern: internal + "/**",
			want:    underInternal,
		},
		{
			name:    "both separators before double star",
			pattern: internal + string(filepath.Separator) + "/**",
			want:    underInternal,
		},
		{
			name:    "double star between segments",
			pattern: app + "/**/services",
			want:    []string{servicesDir, foo},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := glob(tt.pattern)
			require.NoError(t, err)
			assert.ElementsMatch(t, cleanPaths(tt.want), cleanPaths(got))
		})
	}
}

func cleanPaths(paths []string) []string {
	cleaned := make([]string, len(paths))
	for i, p := range paths {
		cleaned[i] = filepath.Clean(p)
	}

	return cleaned
}
