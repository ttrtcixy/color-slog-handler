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

type jsonBuilder struct {
}

func NewJsonHandler(w io.Writer, cfg *Config) *Handler {
	if w == nil {
		w = os.Stderr
	}

	if cfg == nil {
		cfg = &Config{Level: 0, BufferedOutput: false}
	}

	handler := newHandler(w, slog.Level(cfg.Level), &jsonBuilder{})

	if cfg.BufferedOutput {
		handler.shared.bw = bufio.NewWriterSize(w, writerBufSize)
		// Start a background routine to periodically flush the buffer.
		// This ensures logs appear even during low activity periods.
		go handler.flusher()
	}

	return handler
}

func (b *jsonBuilder) buildLog(buf []byte, record slog.Record, precomputedAttrs string, groupPrefix string) []byte {
	buf = append(buf, `{"time":"`...)
	buf = record.Time.AppendFormat(buf, time.DateTime)
	buf = append(buf, `","level":"`...)
	buf = append(buf, levelBytes(record.Level)...)
	buf = append(buf, `","msg":"`...)
	buf = append(buf, record.Message...)
	buf = append(buf, '"')

	if record.NumAttrs() > 0 || precomputedAttrs != "" {
		buf = append(buf, ',')
		if groupPrefix != "" {
			buf = append(buf, groupPrefix...)
		}

		if record.NumAttrs() > 0 {

			if precomputedAttrs != "" {
				buf = append(buf, precomputedAttrs...)
				buf = append(buf, ',')
			}

			var isFirst = true
			record.Attrs(func(attr slog.Attr) bool {
				//attr.Value = attr.Value.Resolve()
				//if attr.Equal(slog.Attr{}) {
				//	return true
				//}

				if !isFirst {
					buf = append(buf, ',')
				} else {
					isFirst = false
				}
				buf = b.appendAttr(buf, nil, attr)
				return true
			})
		} else {
			buf = append(buf, precomputedAttrs...)
		}

		if groupPrefix != "" {
			buf = append(buf, '}')
		}
	}

	buf = append(buf, '}', '\n')
	return buf
}

func (b *jsonBuilder) appendAttr(buf []byte, _ []byte, attr slog.Attr) []byte {
	//attr.Value = attr.Value.Resolve()

	if attr.Equal(slog.Attr{}) {
		return buf
	}

	// Handle nested groups by recursion.
	if attr.Value.Kind() == slog.KindGroup {
		group := attr.Value.Group()

		// If no attrs in group - slog.Group("group",).
		if len(group) == 0 {
			return buf
		}

		if attr.Key != "" {
			buf = append(buf, '"')
			buf = append(buf, attr.Key...)
			buf = append(buf, `":{`...)
		}

		var isFirst = true
		for _, v := range group {
			if v.Equal(slog.Attr{}) {
				continue
			}

			if !isFirst {
				buf = append(buf, ',')
			} else {
				isFirst = false
			}
			buf = b.appendAttr(buf, nil, v)
		}

		if attr.Key != "" {
			buf = append(buf, '}')
		}
		return buf
	}

	buf = append(buf, '"')
	if attr.Key == "" {
		buf = append(buf, "!EMPTY_KEY"...)
	} else {
		buf = append(buf, attr.Key...)
	}
	buf = append(buf, `":`...)
	buf = b.writeValue(buf, attr.Value)

	return buf
}

func (b *jsonBuilder) writeValue(buf []byte, value slog.Value) []byte {
	switch value.Kind() {
	case slog.KindString:
		str := value.String()

		buf = append(buf, '"')
		if str == "" {
			buf = append(buf, "!EMPTY_VALUE"...)
		} else {
			//buf = strconv.AppendQuote(buf, str)
			buf = append(buf, str...)

		}
		buf = append(buf, '"')
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
		buf = strconv.AppendInt(buf, value.Duration().Nanoseconds(), 10)
	case slog.KindTime:
		buf = append(buf, '"')
		buf = value.Time().AppendFormat(buf, time.DateTime)
		buf = append(buf, '"')
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

func (b *jsonBuilder) precomputeAttrs(buf []byte, _ string, attrs []slog.Attr) []byte {
	var attrsCount = len(attrs) - 1

	for i, attr := range attrs {
		buf = b.appendAttr(buf, nil, attr)

		if attrsCount != i {
			buf = append(buf, ',')
		}
	}

	return buf
}

func (b *jsonBuilder) groupPrefix(oldPrefix string, newPrefix string) string {
	return oldPrefix + `"` + newPrefix + `":{`
}
