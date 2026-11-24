package processor

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestAnalyzeFile(t *testing.T) {
	path := filepath.Join("..", "testdata", "orders-a.txt")
	got := analyzeFile(path)
	if got.Err != nil || got.Path != path || got.Lines != 3 || got.Words != 5 {
		t.Fatalf("результат = %+v, ожидаем этот путь, 3 строки и 5 слов", got)
	}
}

func TestProcessFiles(t *testing.T) {
	paths := []string{
		filepath.Join("..", "testdata", "orders-a.txt"),
		filepath.Join("..", "testdata", "orders-b.txt"),
		filepath.Join("..", "testdata", "orders-c.txt"),
		filepath.Join("..", "testdata", "empty.txt"),
	}
	want := map[string][2]int{
		paths[0]: {3, 5},
		paths[1]: {2, 4},
		paths[2]: {13, 12},
		paths[3]: {0, 0},
	}

	for got := range ProcessFiles(paths, 2) {
		stats, found := want[got.Path]
		if !found {
			t.Fatalf("лишний результат для %q", got.Path)
		}
		if got.Err != nil || got.Lines != stats[0] || got.Words != stats[1] {
			t.Fatalf("для %q получили %+v, ожидаем строк и слов %v", got.Path, got, stats)
		}
		delete(want, got.Path)
	}
	if len(want) > 0 {
		t.Fatalf("не получены результаты: %v", want)
	}
}

func TestNonExistingFiles(t *testing.T) {
	paths := []string{
		filepath.Join("..", "testdata", "not-existing.txt"),
	}

	for got := range ProcessFiles(paths, 2) {
		if got.Err == nil || got.Path != paths[0] {
			t.Fatalf("non empty result for non existing file")
		}
		if !errors.Is(got.Err, fs.ErrNotExist) {
			t.Fatalf("for %q want fs.ErrNotExits, got %+v", got.Path, got.Err)
		}
	}
}
