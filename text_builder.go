package logger

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"
)

var (
	reset   = []byte("\033[0m")
	red     = []byte("\033[31m")
	green   = []byte("\033[32m")
	yellow  = []byte("\033[33m")
	blue    = []byte("\033[34m")
	magenta = []byte("\033[35m")
	cyan    = []byte("\033[36m")
	none    = []byte("")
)

type colorOptions struct {
	TimeColor  []byte
	KeyColor   []byte
	ValueColor []byte
}

func newColorOptions(timeColor, keyColor, valueColor []byte) *colorOptions {
	return &colorOptions{
		TimeColor:  timeColor,
		KeyColor:   keyColor,
		ValueColor: valueColor,
	}
}

type colorizedTextBuilder struct {
	colorOpts *colorOptions
}

func NewTextHandler(w io.Writer, cfg *Config) *Handler {
	if w == nil {
		w = os.Stderr
	}

	if cfg == nil {
		cfg = &Config{Level: 0, BufferedOutput: false}
	}

	textBuilder := &colorizedTextBuilder{
		colorOpts: newColorOptions(blue, magenta, none),
	}

	handler := newHandler(w, slog.Level(cfg.Level), textBuilder)

	if cfg.BufferedOutput {
		handler.shared.bw = bufio.NewWriterSize(w, writerBufSize)
		// Start a background routine to periodically flush the buffer.
		// This ensures logs appear even during low activity periods.
		go handler.flusher()
	}

	return handler
}

func (b *colorizedTextBuilder) buildLog(buf []byte, record slog.Record, precomputedAttrs string, groupPrefix string) []byte {
	// Time
	buf = append(buf, b.colorOpts.TimeColor...) // color
	buf = record.Time.AppendFormat(buf, time.TimeOnly)
	buf = append(buf, reset...) // color
	buf = append(buf, " | "...)

	// Level
	levelColor := levelColor(record.Level)
	buf = append(buf, levelColor...) // color
	buf = append(buf, levelBytes(record.Level)[:4]...)
	buf = append(buf, reset...) // color

	buf = append(buf, " | "...)

	// Message
	buf = append(buf, levelColor...) // color
	buf = append(buf, record.Message...)
	buf = append(buf, reset...) // color

	// Append precomputed attributes (from WithAttrs)
	if len(precomputedAttrs) > 0 {
		buf = append(buf, precomputedAttrs...)
	}
	// Process dynamic attributes (attached to this specific record)
	if record.NumAttrs() > 0 {
		// Stack-allocated buffer for group prefix to avoid allocs
		var groupBuf [128]byte
		pref := groupBuf[:0]

		// Add group from WithGroup()
		if len(groupPrefix) > 0 {
			pref = append(pref, groupPrefix...)
		}

		record.Attrs(func(attr slog.Attr) bool {
			buf = b.appendAttr(buf, pref, attr)
			return true
		})
	}

	buf = append(buf, '\n')
	return buf
}

func (b *colorizedTextBuilder) appendAttr(buf []byte, groupPrefix []byte, attr slog.Attr) []byte {
	attr.Value = attr.Value.Resolve()

	if attr.Equal(slog.Attr{}) {
		return buf
	}

	// Handle nested groups by recursion: flattening keys to "prefix.key"
	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			groupPrefix = append(groupPrefix, attr.Key...)
			groupPrefix = append(groupPrefix, '.')
		}

		for _, v := range attr.Value.Group() {
			buf = b.appendAttr(buf, groupPrefix, v)
		}
		return buf
	}

	buf = append(buf, ' ')
	buf = append(buf, b.colorOpts.KeyColor...) // color

	if len(groupPrefix) > 0 {
		buf = append(buf, groupPrefix...)
	}

	if attr.Key == "" {
		attr.Key = "!EMPTY_KEY"
	}
	buf = append(buf, attr.Key...)
	buf = append(buf, '=')
	buf = append(buf, reset...) // color

	buf = append(buf, b.colorOpts.ValueColor...) // color
	buf = b.writeValue(buf, attr.Value)
	buf = append(buf, reset...) // color

	return buf
}

func (b *colorizedTextBuilder) writeValue(buf []byte, value slog.Value) []byte {
	switch value.Kind() {
	case slog.KindString:
		str := value.String()
		if str == "" {
			str = "!EMPTY_VALUE"
		}
		buf = append(buf, str...)
	case slog.KindInt64:
		buf = strconv.AppendInt(buf, value.Int64(), 10)
	case slog.KindUint64:
		buf = strconv.AppendUint(buf, value.Uint64(), 10)
	case slog.KindFloat64:
		buf = strconv.AppendFloat(buf, value.Float64(), 'f', -1, 64)
	case slog.KindBool:
		if value.Bool() {
			buf = append(buf, "true"...)
		} else {
			buf = append(buf, "false"...)
		}
	case slog.KindDuration:
		buf = append(buf, value.Duration().String()...)
	case slog.KindTime:
		buf = value.Time().AppendFormat(buf, time.DateTime)
	case slog.KindAny:
		//if err, ok := value.Any().(error); ok {
		//	buf = append(buf, err.Error()...)
		//	return buf
		//}
		b, err := json.Marshal(value.Any())
		if err != nil {
			buf = append(buf, "!ERR_MARSHAL"...)
		} else {
			buf = append(buf, b...)
		}
	default:
		buf = append(buf, "!UNHANDLED"...)
	}
	return buf
}

func (b *colorizedTextBuilder) precomputeAttrs(buf []byte, groupPrefix string, attrs []slog.Attr) []byte {
	// Prepare the current group prefix for these specific attributes.
	var groupBuf [128]byte
	pref := groupBuf[:0]

	// Add group from WithGroup()
	if len(groupPrefix) > 0 {
		pref = append(pref, groupPrefix...)
	}

	for _, attr := range attrs {
		buf = b.appendAttr(buf, pref, attr)
	}

	return buf
}

func (b *colorizedTextBuilder) groupPrefix(oldPrefix string, newPrefix string) string {
	return oldPrefix + newPrefix + "."
}
