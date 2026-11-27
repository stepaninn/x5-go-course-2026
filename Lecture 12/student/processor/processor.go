package processor

import (
	"context"
	"sync"
	"time"
)

// ProcessFiles запускает обработчик и сразу возвращает два канала.
// results передаёт результаты; done закрывается после выхода раздатчика и работников.
// После отмены получатель может перестать читать results, но должен дождаться done.
func ProcessFiles(ctx context.Context, paths []string, workers int, lineDelay time.Duration) (<-chan Result, <-chan struct{}) {
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan string)
	results := make(chan Result)
	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobs)
		for _, path := range paths {
			// TODO 2: дайте раздатчику выйти, если контекст отменили.
			jobs <- path
		}
	}()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO 2: работник должен замечать отмену и при ожидании задания.
			for path := range jobs {
				// TODO 2: не начинайте новый файл, если отмена уже известна.
				result := analyzeFile(ctx, path, lineDelay)
				// TODO 2: получатель после отмены может больше не читать results.
				results <- result
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
		close(done)
	}()
	return results, done
}
