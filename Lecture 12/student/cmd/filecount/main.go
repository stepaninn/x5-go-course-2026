package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"example.com/lesson12/files/processor"
)

func main() {
	budget := flag.Duration("timeout", 300*time.Millisecond, "общий бюджет обработки")
	lineDelay := flag.Duration("line-delay", 0, "имитация работы над каждой строкой")
	workers := flag.Int("workers", 3, "число работников")
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"testdata/orders-a.txt", "testdata/orders-b.txt", "testdata/empty.txt"}
	}
	fmt.Println("Заданный бюджет:", *budget)
	// TODO 1: создайте один контекст с этим бюджетом на весь прогон.
	// Освободите его ресурсы при выходе из main.
	ctx := context.Background()
	start := time.Now()
	results, done := processor.ProcessFiles(ctx, paths, *workers, *lineDelay)
	count := 0
	for result := range results {
		count++
		if result.Err != nil {
			fmt.Printf("%s: %v\n", result.Path, result.Err)
			continue
		}
		fmt.Printf("%s: строк %d, слов %d\n", result.Path, result.Lines, result.Words)
	}
	<-done
	fmt.Println("Получено результатов:", count)
	fmt.Println("Все работники и раздатчик завершились")
	fmt.Println("Заняло:", time.Since(start).Round(time.Millisecond))
	fmt.Println("Причина завершения:", ctx.Err())
}
