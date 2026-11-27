package processor

import (
	"bufio"
	"context"
	"errors"
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
	// TODO 1: не открывайте новый файл, если ctx уже отменён.
	file, err := os.Open(path)
	if err != nil {
		return Result{Path: path, Err: err}
	}
	defer file.Close()

	result := Result{Path: path}
	reader := bufio.NewReader(file)
	for {
		// TODO 1: проверяйте отмену между чтениями строк.
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
	// TODO 1: замените Sleep ожиданием, которое реагирует на ctx.Done().
	// Если ctx уже отменён, верните его ошибку даже при delay == 0.
	time.Sleep(delay)
	return nil
}
