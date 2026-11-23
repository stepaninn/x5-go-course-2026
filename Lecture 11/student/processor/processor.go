package processor

import (
	"fmt"
	"sync"
)

// ProcessFiles возвращает по результату на каждый переданный путь.
// Получатель читает канал до закрытия.
func ProcessFiles(paths []string, workers int) <-chan Result {
	if workers < 1 {
		workers = 1
	}
	results := make(chan Result)
	// TODO: распределите пути между workers обработчиками через канал jobs.
	// Для каждого пути вызовите analyzeFile и отправьте результат в results.
	// Определите, кто закроет каждый канал и как для этого использовать WaitGroup.

	jobs := make(chan Result)
	close(jobs)

	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				results <- analyzeFile(file.Path)
			}
			fmt.Printf("worker %d done\n", w)
		}()
	}

	wg.Wait()
	close(results)

	return results
}
