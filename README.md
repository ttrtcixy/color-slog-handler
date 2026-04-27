[English](README.md) | [Русский](README.ru.md)

# Fast `slog` Handlers

Lightweight and optimized handlers for Go `log/slog` with a focus on throughput and low allocations.

- `NewJsonHandler`: compact JSON output for production.
- `NewTextHandler`: readable ANSI-colored text output for local development.

## Features

- Low-allocation formatting with `sync.Pool`.
- Optional buffered output (`bufio.Writer`) with periodic background flushing.
- Context attributes support via `AppendAttrsToCtx(...)`.
- Safe concurrent use.
- Supports `WithAttrs` and `WithGroup`.

## Install

```shell
go get github.com/ttrtcixy/fast-slog-handler
```

## Quick Start

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
		WriteBuffSize:  4096,               // optional
		FlushInterval:  250 * time.Millisecond, // optional
	}

	handler := logger.NewJsonHandler(os.Stdout, cfg) // or NewTextHandler for local dev
	log := slog.New(handler)

	// Required only for buffered mode.
	defer func() { _ = handler.Close(context.Background()) }()

	ctx := logger.AppendAttrsToCtx(
		context.Background(),
		slog.String("trace_id", "af82-bx22"),
		slog.String("request_id", "req-42"),
	)

	log.LogAttrs(ctx, slog.LevelInfo, "payment accepted", slog.Int("amount", 500))
}
```

## Configuration

`Config` fields:

- `Level` (`slog.Level`): minimum enabled log level.
- `BufferedOutput` (`bool`): enable buffered writes and background flusher.
- `WriteBuffSize` (`int`): writer buffer size, default `4096`.
- `FlushInterval` (`time.Duration`): flush interval, default `250ms`.
- `MaxBufPoolSize` (`int`): max pooled formatter buffer size, default `4096`.

## Buffering and `Close()`

When `BufferedOutput` is enabled, call `Close(ctx)` before shutdown to:

- stop the flusher goroutine;
- flush remaining data from the write buffer.

If buffering is disabled, `Close()` returns `ErrNothingToClose`.
