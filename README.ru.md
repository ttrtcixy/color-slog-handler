[English](README.md) | [Русский](README.ru.md)

# Быстрые `slog`-обработчики

Легковесные и оптимизированные обработчики для Go `log/slog` с фокусом на производительность и низкие аллокации.

- `NewJsonHandler`: компактный JSON-вывод для production.
- `NewTextHandler`: читаемый текстовый вывод с ANSI-подсветкой для локальной разработки.

## Что есть

- Низкоаллокирующее форматирование через `sync.Pool`.
- Опциональный буферизированный вывод (`bufio.Writer`) с периодическим фоновым flush.
- Контекстные атрибуты через `AppendAttrsToCtx(...)`.
- Потокобезопасная работа.
- Поддержка `WithAttrs` и `WithGroup`.

## Установка

```shell
go get github.com/ttrtcixy/color-slog-handler
```

## Быстрый старт

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	logger "github.com/ttrtcixy/color-slog-handler"
)

func main() {
	cfg := &logger.Config{
		Level:          slog.LevelInfo,
		BufferedOutput: true,
		WriteBuffSize:  4096,               // опционально
		FlushInterval:  250 * time.Millisecond, // опционально
	}

	handler := logger.NewJsonHandler(os.Stdout, cfg) // или NewTextHandler для локальной разработки
	log := slog.New(handler)

	// Нужен только при BufferedOutput=true.
	defer func() { _ = handler.Close(context.Background()) }()

	ctx := logger.AppendAttrsToCtx(
		context.Background(),
		slog.String("trace_id", "af82-bx22"),
		slog.String("request_id", "req-42"),
	)

	log.LogAttrs(ctx, slog.LevelInfo, "payment accepted", slog.Int("amount", 500))
}
```

## Конфигурация

Поля `Config`:

- `Level` (`slog.Level`): минимальный уровень логирования.
- `BufferedOutput` (`bool`): включить буферизированную запись и фоновый flusher.
- `WriteBuffSize` (`int`): размер буфера записи, по умолчанию `4096`.
- `FlushInterval` (`time.Duration`): интервал flush, по умолчанию `250ms`.
- `MaxBufPoolSize` (`int`): максимальный размер буфера форматтера в пуле, по умолчанию `4096`.

## Буферизация и `Close()`

Если включен `BufferedOutput`, перед завершением процесса вызывайте `Close(ctx)`, чтобы:

- остановить фоновую goroutine flush;
- сбросить оставшиеся данные из write-буфера.

Если буферизация выключена, `Close()` вернет `ErrNothingToClose`.
