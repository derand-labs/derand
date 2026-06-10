package sol

import (
	"fmt"
	"strings"
)

func solcErrorsString(errs []map[string]any) string {
	s := []string{}
	for _, e := range errs {
		s = append(s, errString(e))
	}

	return strings.Join(s, "\n")
}

func errString(e map[string]any) string {
	sev := get[string](e, "severity")
	code := get[string](e, "errorCode")
	fmsg := strings.ReplaceAll(strings.TrimSpace(get[string](e, "formattedMessage")), "\n", " ")
	return fmt.Sprintf("[%s-%s] %s", sev, code, fmsg)
}

func get[T any](m map[string]any, key string) T {
	v, ok := m[key]
	if !ok {
		var zero T
		return zero
	}

	val, ok := v.(T)
	if !ok {
		var zero T
		return zero
	}

	return val
}
