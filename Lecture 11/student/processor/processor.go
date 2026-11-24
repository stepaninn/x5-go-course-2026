package processor

import (
	"sync"
)

type TotalCount struct {
	TotalLines int
	TotalWords int
}

// ProcessFiles возвращает по результату на каждый переданный путь.
// Получатель читает канал до закрытия.
func ProcessFiles(paths []string, workers int) (<-chan Result, <-chan TotalCount) {
	if workers < 1 {
		workers = 1
	}
	results := make(chan Result)
	jobs := make(chan Result)

	totals := make(chan TotalCount, 1)
	totalLines := 0
	totalWords := 0
	mu := sync.Mutex{}

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
				result := analyzeFile(file.Path)

				if result.Err == nil {
					mu.Lock()
					totalLines += result.Lines
					totalWords += result.Words
					mu.Unlock()
				}

				results <- result
			}
		}()
	}

	go func() {
		wg.Wait()
		totals <- TotalCount{TotalLines: totalLines, TotalWords: totalWords}
		close(totals)
		close(results)
	}()

	return results, totals
}
