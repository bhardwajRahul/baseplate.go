package log_test

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/reddit/baseplate.go/log"
	"gopkg.in/yaml.v2"
)

func TestLogWrapperNilSafe(t *testing.T) {
	// Just make sure log.Wrapper.Log is nil-safe, no real tests
	var logger log.Wrapper
	logger.Log(context.Background(), "Hello, world!")
	logger.ToThriftLogger()("Hello, world!")
	log.WrapToThriftLogger(nil)("Hello, world!")
}

func TestLogWrapperUnmarshalText(t *testing.T) {
	getActualFuncName := func(w log.Wrapper) string {
		// This returns something like:
		// "github.com/reddit/baseplate.go/log.ZapWrapper.func1"
		return runtime.FuncForPC(reflect.ValueOf(w).Pointer()).Name()
	}

	for _, c := range []struct {
		text     string
		err      bool
		expected string
	}{
		{
			text: "fancy",
			err:  true,
		},
		{
			text:     "",
			expected: "ErrorWithSentryWrapper",
		},
		{
			text:     "nop",
			expected: "NopWrapper",
		},
		{
			text:     "std",
			expected: "StdWrapper",
		},
		{
			text:     "zap",
			expected: "ZapWrapper",
		},
		{
			// Unfortunately there's no way to check that the arg passed into
			// ZapWrapper is correct.
			text:     "zap:error",
			expected: "ZapWrapper",
		},
		{
			text: "zap:error:key",
			err:  true, // expect error because of dangling key.
		},
		{
			text: "zap:error:key=value:extra",
			err:  true, // expect error because of extra.
		},
		{
			text:     "zap:info:key1=value1,key2=value2 with space",
			expected: "ZapWrapper",
		},
		{
			text: "zaperror",
			err:  true,
		},
		{
			text:     "sentry",
			expected: "ErrorWithSentryWrapper",
		},
	} {
		t.Run(c.text, func(t *testing.T) {
			var w log.Wrapper
			err := w.UnmarshalText([]byte(c.text))
			if c.err {
				if err == nil {
					t.Errorf(
						"Expected UnmarshalText to return error, got nil. Result is %q",
						getActualFuncName(w),
					)
				}
			} else {
				if err != nil {
					t.Errorf("Expected UnmarshalText to return nil error, got %v", err)
				}
				name := getActualFuncName(w)
				if !strings.Contains(name, c.expected) {
					t.Errorf("Expected function name to contain %q, got %q", c.expected, name)
				}
			}
		})
	}
}

func TestWrapperMarshalYAML(t *testing.T) {
	type foo struct {
		Log log.Wrapper `yaml:"log,omitempty"`
	}
	t.Run("nil", func(t *testing.T) {
		bar := foo{Log: nil}
		var sb strings.Builder
		if err := yaml.NewEncoder(&sb).Encode(bar); err != nil {
			t.Fatal(err)
		}
		if got, want := strings.TrimSpace(sb.String()), `log: null`; got != want {
			t.Errorf("yaml got %q want %q", got, want)
		}
	})
	t.Run("non-nil", func(t *testing.T) {
		bar := foo{Log: log.NopWrapper}
		var sb strings.Builder
		if err := yaml.NewEncoder(&sb).Encode(bar); err == nil {
			t.Errorf("Expected marshal error, got yaml %q", sb.String())
		}
	})
}

var (
	_ yaml.Marshaler = (log.Wrapper)(nil)
)

var (
	_ log.Counter = (prometheus.Counter)(nil)
	_ log.Counter = (metrics.Counter)(nil)
)
