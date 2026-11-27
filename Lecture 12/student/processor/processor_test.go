package processor

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func fixture(name string) string { return filepath.Join("..", "testdata", name) }

func TestAnalyzeFile(t *testing.T) {
	got := analyzeFile(context.Background(), fixture("orders-a.txt"), 0)
	if got.Err != nil || got.Lines != 3 || got.Words != 5 {
		t.Fatalf("получили %+v, нужны 3 строки и 5 слов без ошибки", got)
	}
}

func TestProcessFiles(t *testing.T) {
	paths := []string{fixture("orders-a.txt"), fixture("orders-b.txt"), fixture("empty.txt")}
	want := map[string][2]int{paths[0]: {3, 5}, paths[1]: {2, 4}, paths[2]: {0, 0}}
	results, done := ProcessFiles(context.Background(), paths, 3, 0)
	for got := range results {
		stats, ok := want[got.Path]
		if !ok || got.Err != nil || got.Lines != stats[0] || got.Words != stats[1] {
			t.Errorf("неожиданный результат: %+v", got)
		}
		delete(want, got.Path)
	}
	awaitDone(t, done)
	if len(want) != 0 {
		t.Errorf("нет результатов: %v", want)
	}
}

func TestDuplicatesAndMissingFile(t *testing.T) {
	path := fixture("orders-a.txt")
	missing := filepath.Join(t.TempDir(), "missing.txt")
	results, done := ProcessFiles(context.Background(), []string{path, missing, path}, 0, 0)
	counts := map[string]int{}
	for got := range results {
		counts[got.Path]++
		if (got.Path == missing) != (got.Err != nil) {
			t.Errorf("неверная ошибка: %+v", got)
		}
	}
	awaitDone(t, done)
	if counts[path] != 2 || counts[missing] != 1 {
		t.Errorf("число результатов: %v", counts)
	}
}

func TestEmptyInput(t *testing.T) {
	results, done := ProcessFiles(context.Background(), nil, 3, 0)
	for result := range results {
		t.Errorf("лишний результат: %+v", result)
	}
	awaitDone(t, done)
}

func TestAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, done := ProcessFiles(ctx, []string{fixture("orders-a.txt")}, 2, 0)
	for result := range results {
		t.Errorf("для заранее отменённого контекста не нужно начинать файл: %+v", result)
	}
	awaitDone(t, done)
}

func TestWaitLineCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := waitLine(ctx, 400*time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ошибка = %v, нужна context.Canceled", err)
	}
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Errorf("заранее отменённая работа ждала %v", elapsed)
	}
}

func TestDeadlineDuringWork(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	results, done := ProcessFiles(ctx, []string{fixture("orders-b.txt")}, 1, 400*time.Millisecond)
	drained := make(chan struct{})
	go func() {
		for range results {
		}
		close(drained)
	}()
	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Error("дедлайн истёк, а работа всё ещё продолжается")
	}
	// Даже исходную заготовку дожидаемся: её имитация работы конечна.
	awaitDone(t, done)
	awaitDone(t, drained)
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("причина = %v, нужна DeadlineExceeded", ctx.Err())
	}
}

func TestCancelWithoutReading(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	paths := make([]string, 100)
	for i := range paths {
		paths[i] = fixture("orders-a.txt")
	}
	results, done := ProcessFiles(ctx, paths, 2, 0)
	// Первый ответ подтверждает, что конвейер уже начал работать.
	select {
	case <-results:
	case <-time.After(time.Second):
		t.Fatal("не получили первый результат")
	}
	cancel()
	// Больше результатов не читаем. Все запущенные функции должны выйти.
	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Error("после отмены кто-то остался ждать читателя results или работника jobs")
		// Освобождаем отправителей в исходной заготовке, чтобы тест не завис.
		for range results {
		}
		awaitDone(t, done)
	}
}

func awaitDone(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("не дождались завершения за 2 секунды")
	}
}
