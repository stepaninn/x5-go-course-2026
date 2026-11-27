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
			select {
			case jobs <- path:
			case <-ctx.Done():
				return
			}
		}
	}()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case path, ok := <-jobs:
					if !ok {
						return
					}
					if err := ctx.Err(); err != nil {
						return
					}

					result := analyzeFile(ctx, path, lineDelay)

					select {
					case <-ctx.Done():
						return
					case results <- result:
					}
				}
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
