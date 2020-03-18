package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	serrors "github.com/go-errors/errors"
	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	log "github.com/sirupsen/logrus"
	"strings"
)

type UtilsFormatter struct {
	log.TextFormatter
}

func (f *UtilsFormatter) Format(entry *log.Entry) ([]byte, error) {
	var entryError error
	for k, v := range entry.Data {
		if k == "error" {
			if unpackedErr, ok := v.(error); ok {
				entryError = unpackedErr
				delete(entry.Data, "error")
			}
		}
	}

	result, err := f.TextFormatter.Format(entry)
	b := bytes.NewBuffer(result)
	openb := bytes.Buffer{}
	// extract opentracing span
	var span opentracing.Span
	if ctx := entry.Context; ctx != nil {
		if span = opentracing.SpanFromContext(entry.Context); span == nil {
			if gctx, ok := ctx.(*gin.Context); ok {
				span = opentracing.SpanFromContext(gctx.Request.Context())
			}
		}
	}
	if entryError != nil {

		if _, ok := entryError.(*serrors.Error); ok {
			b.WriteString(fmt.Sprint("error: \n"))
			var stackError = entryError
			for stackError != nil {
				if stackUnwrapError, ok := stackError.(*serrors.Error); ok {
					stack := stackUnwrapError.ErrorStack()
					stackStrings := strings.Split(stack, "\n")
					for i, v := range stackStrings[1:] {
						stackStrings[i+1] = "\t" + v
					}

					if log.GetLevel() < log.DebugLevel && len(stackStrings) > 11 {
						_stackStrings := stackStrings[:11]
						stack = strings.Join(_stackStrings, "\n")
						b.WriteString(fmt.Sprintf("- %s\n\t<skip in non debugger level mode> ...\n", stack))
					} else {
						stack = strings.Join(stackStrings, "\n")
						b.WriteString(fmt.Sprintf("- %+v\n", stack))
					}
					if span != nil {
						stack = strings.Join(stackStrings, "\n")
						openb.WriteString(fmt.Sprintf("- %+v\n", stack))
					}
					if stackError != errors.Unwrap(stackUnwrapError.Err) {
						stackError = errors.Unwrap(stackUnwrapError.Err)
					} else {
						stackError = nil
					}
				} else {
					b.WriteString(fmt.Sprintf("- %+v\n", stackError))
					stackError = errors.Unwrap(stackError)
				}

			}

		} else {
			b.WriteString(fmt.Sprintf("error: %+v \n", entryError))
		}

	}
	result = b.Bytes()
	//extract error and fields to OpenTracing log

	if span != nil {
		span.SetTag("error", true)
		fields := []olog.Field{
			olog.String("event", "error"),
			olog.String("error.message", entry.Message),
			olog.Error(entryError),
			olog.String("error.stacktrace", string(openb.Bytes())),
		}
		if entry.HasCaller() {
			fields = append(fields, olog.String("error.caller", fmt.Sprintf("%s:%d %s", entry.Caller.File, entry.Caller.Line, entry.Caller.Function)))
		}
		for k, v := range entry.Data {
			fields = append(fields, olog.Object("error."+k, v))

		}
		span.LogFields(
			fields...,
		)
	}

	return result, err
}
