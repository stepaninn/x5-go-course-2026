package processor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Result struct {
	Path  string
	Lines int
	Words int
	Err   error
}

// analyzeFile считает строки и слова, как на занятии 11.
// lineDelay имитирует работу над строкой, например проверку заказов.
func analyzeFile(ctx context.Context, path string, lineDelay time.Duration) Result {
	if err := ctx.Err(); err != nil {
		return Result{Path: path, Err: err}
	}
	file, err := os.Open(path)
	if err != nil {
		return Result{Path: path, Err: err}
	}
	defer file.Close()

	result := Result{Path: path}
	reader := bufio.NewReader(file)
	for {
		if err := ctx.Err(); err != nil {
			return Result{Path: path, Err: err}
		}
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			if err := waitLine(ctx, lineDelay); err != nil {
				return Result{Path: path, Err: err}
			}
			result.Lines++
			result.Words += len(strings.Fields(line))
		}
		if errors.Is(err, io.EOF) {
			return result
		}
		if err != nil {
			return Result{Path: path, Err: err}
		}
	}
}

func waitLine(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		fmt.Println("строка обработана")
		return nil
	}
}
