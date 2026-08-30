package logging

import (
	"fmt"
	"log/slog"
	"path"
	"reflect"
	"strings"
	"time"

	sloggin "github.com/samber/slog-gin"
)

func ParseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func NewStructuredLoggingConfig() *sloggin.Config {
	config := sloggin.DefaultConfig()
	config.WithUserAgent = true
	config.WithClientIP = true
	config.Filters = append(config.Filters, sloggin.IgnorePath("/healthz", "/metrics"))
	return &config
}

// ReplaceAttr is a slog.HandlerOptions.ReplaceAttr function that
// recursively processes attributes to ensure time.Duration and errors are formatted correctly.
func ReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		if source, ok := a.Value.Any().(*slog.Source); ok {
			file := source.File
			if idx := strings.LastIndex(file, "/internal/"); idx != -1 {
				file = file[idx+1:]
			} else {
				file = path.Base(file)
			}
			return slog.String("caller", fmt.Sprintf("%s:%d", file, source.Line))
		}
	}

	if a.Value.Kind() != slog.KindAny {
		return a
	}

	val := a.Value.Any()
	processed := processValue(val, 0)

	return slog.Any(a.Key, processed)
}

func processValue(v any, depth int) any {
	if depth > 10 {
		return v
	}

	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return v
	}

	// Handle pointers by dereferencing
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	switch vTyped := v.(type) {
	case time.Duration:
		return vTyped.String()
	case error:
		return v
	}

	switch val.Kind() {
	case reflect.Struct:
		t := val.Type()
		newMap := make(map[string]any, val.NumField())
		for i := 0; i < val.NumField(); i++ {
			fieldType := t.Field(i)
			if !fieldType.IsExported() {
				continue
			}

			name := fieldType.Name
			tag := fieldType.Tag.Get("json")
			if tag == "-" {
				continue
			}
			if tag != "" {
				if idx := strings.Index(tag, ","); idx != -1 {
					tag = tag[:idx]
				}
				if tag != "" {
					name = tag
				}
			}

			fieldVal := val.Field(i).Interface()
			if fieldType.Tag.Get("log") == "redact" {
				newMap[name] = "********"
			} else {
				newMap[name] = processValue(fieldVal, depth+1)
			}
		}
		return newMap
	case reflect.Map:
		newMap := make(map[string]any, val.Len())
		for _, key := range val.MapKeys() {
			mapKeyStr := fmt.Sprintf("%v", key.Interface())
			mapVal := val.MapIndex(key).Interface()
			newMap[mapKeyStr] = processValue(mapVal, depth+1)
		}
		return newMap
	case reflect.Slice, reflect.Array:
		newSlice := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			newSlice[i] = processValue(val.Index(i).Interface(), depth+1)
		}
		return newSlice
	default:
		return v
	}
}
