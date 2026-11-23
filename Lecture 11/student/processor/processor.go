package processor

import (
	"sync"
)

// ProcessFiles возвращает по результату на каждый переданный путь.
// Получатель читает канал до закрытия.
func ProcessFiles(paths []string, workers int) <-chan Result {
	if workers < 1 {
		workers = 1
	}
	results := make(chan Result)
	jobs := make(chan Result)
	go func() {
		for _, path := range paths {
			jobs <- Result{
				Path: path,
			}
		}
		close(jobs)
	}()

	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				results <- analyzeFile(file.Path)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
