package scanner_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/willie68/go-arch-lint/internal/models"
	"github.com/willie68/go-arch-lint/internal/services/project/scanner"
)

// TestScan_IgnoresUnstatableOutOfScopeFile covers a directory that can be listed
// but whose entries cannot be stat'ed (mode 0o400: read, no search). The files
// inside are out of scope anyway (not ".go"), so the scan must not care about
// them. Walking with filepath.Walk stat'ed every entry before the scope check
// and aborted the whole scan on the first permission error; walking with
// filepath.WalkDir never stats out-of-scope entries.
func TestScan_IgnoresUnstatableOutOfScopeFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions do not restrict stat on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses directory permissions; cannot simulate an unstatable entry")
	}

	projectDir := t.TempDir()

	goFile := filepath.Join(projectDir, "main.go")
	if err := os.WriteFile(goFile, []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	// Not excluded in the config: the walk descends into it and lists it fine,
	// but stat'ing what it contains is denied.
	unsearchableDir := filepath.Join(projectDir, "assets")
	if err := os.Mkdir(unsearchableDir, 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(unsearchableDir, ".env"), []byte("SECRET=1\n"), 0o600); err != nil {
		t.Fatalf("write out of scope file: %v", err)
	}
	if err := os.Chmod(unsearchableDir, 0o400); err != nil {
		t.Fatalf("chmod dir: %v", err)
	}
	// Restore perms so t.TempDir cleanup can remove it.
	t.Cleanup(func() { _ = os.Chmod(unsearchableDir, 0o755) })

	files, err := scanner.NewScanner().Scan(context.Background(), projectDir, "example.com/proj", nil, nil)
	if err != nil {
		t.Fatalf("scan should ignore the unstatable out of scope file, got error: %v", err)
	}

	if len(files) != 1 || files[0].Path != goFile {
		t.Fatalf("expected only %q to be scanned, got %+v", goFile, files)
	}
}

// TestScan_SkipsUnreadableExcludedDir reproduces the case where an excluded
// directory inside the project tree is unreadable (e.g. a root-owned local
// docker volume). Before the fix the walk descended into it, hit a
// permission-denied readdir error and aborted the whole scan. Excluded dirs
// must now be skipped before descending, so the scan succeeds and still
// returns the in-scope source files.
func TestScan_SkipsUnreadableExcludedDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses directory permissions; cannot simulate an unreadable dir")
	}

	projectDir := t.TempDir()

	goFile := filepath.Join(projectDir, "main.go")
	if err := os.WriteFile(goFile, []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	excludedDir := filepath.Join(projectDir, ".data")
	if err := os.Mkdir(excludedDir, 0o755); err != nil {
		t.Fatalf("mkdir excluded dir: %v", err)
	}
	// A subdir the walker would try to readdir into; made unreadable so that
	// descending into it fails, mirroring the root-owned postgres data dir.
	unreadable := filepath.Join(excludedDir, "pgdata")
	if err := os.Mkdir(unreadable, 0o000); err != nil {
		t.Fatalf("mkdir unreadable dir: %v", err)
	}
	// Restore perms so t.TempDir cleanup can remove it.
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o755) })

	excludePaths := []models.ResolvedPath{{
		LocalPath: ".data",
		AbsPath:   excludedDir,
	}}

	files, err := scanner.NewScanner().Scan(context.Background(), projectDir, "example.com/proj", excludePaths, nil)
	if err != nil {
		t.Fatalf("scan should skip the excluded unreadable dir, got error: %v", err)
	}

	if len(files) != 1 || files[0].Path != goFile {
		t.Fatalf("expected only %q to be scanned, got %+v", goFile, files)
	}
}

// TestScan_DoesNotDescendIntoExcludedDir asserts that .go files living inside
// an excluded directory are never scanned.
func TestScan_DoesNotDescendIntoExcludedDir(t *testing.T) {
	projectDir := t.TempDir()

	kept := filepath.Join(projectDir, "keep.go")
	if err := os.WriteFile(kept, []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write kept file: %v", err)
	}

	excludedDir := filepath.Join(projectDir, "vendor")
	if err := os.Mkdir(excludedDir, 0o755); err != nil {
		t.Fatalf("mkdir excluded dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(excludedDir, "dep.go"), []byte("package vendor\n"), 0o600); err != nil {
		t.Fatalf("write excluded file: %v", err)
	}

	excludePaths := []models.ResolvedPath{{
		LocalPath: "vendor",
		AbsPath:   excludedDir,
	}}

	files, err := scanner.NewScanner().Scan(context.Background(), projectDir, "example.com/proj", excludePaths, nil)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if len(files) != 1 || files[0].Path != kept {
		t.Fatalf("expected only %q to be scanned, got %+v", kept, files)
	}
}
