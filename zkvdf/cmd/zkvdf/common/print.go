package common

import (
	"fmt"
	"strings"
	"time"

	gnarklogger "github.com/consensys/gnark/logger"
)

const (
	FontBold   = "\033[1m"
	FontGreen  = "\033[32m"
	FontRed    = "\033[31m"
	FontYellow = "\033[33m"
	FontReset  = "\033[0m"
)

func init() {
	gnarklogger.Disable()
}

func PrintWarn(s string, args ...any) {
	fmt.Println(FontYellow + FontBold + "[WARNING] " + FontReset + FontBold + fmt.Sprintf(s, args...) + FontReset)
}

type Step[T1, T2 any] struct {
	name  string
	ok    []string
	okf   []func(T1, T2) string
	fail  []string
	failf []func(error) string
}

func NewStep0(name string) *Step[any, any] {
	return &Step[any, any]{name: name}
}

func NewStep1[T any, S ~string](name S) *Step[T, any] {
	return &Step[T, any]{name: string(name)}
}

func NewStep2[T1, T2 any, S ~string](name S) *Step[T1, T2] {
	return &Step[T1, T2]{name: string(name)}
}

func (s *Step[T1, T2]) OkMessage(msg ...string) *Step[T1, T2] {
	s.ok = append(s.ok, msg...)
	return s
}

func (s *Step[T1, T2]) FailMessage(msg ...string) *Step[T1, T2] {
	s.fail = append(s.fail, msg...)
	return s
}

func (s *Step[T1, T2]) OkMessageFunc1(f func(T1) string) *Step[T1, T2] {
	f2 := func(t1 T1, _ T2) string {
		return f(t1)
	}
	s.okf = append(s.okf, f2)
	return s
}

func (s *Step[T1, T2]) OkMessageFunc2(f func(T1, T2) string) *Step[T1, T2] {
	s.okf = append(s.okf, f)
	return s
}

func (s *Step[T1, T2]) FailMessageFunc(f func(error) string) *Step[T1, T2] {
	s.failf = append(s.failf, f)
	return s
}

func (s *Step[T1, T2]) Do0(f func() error) error {
	printTimerHeading(s.name)

	start := time.Now()
	err := f()
	msgs := []string{}
	if err == nil {
		var t1 T1
		var t2 T2
		msgs = append(msgs, s.ok...)
		for _, fmsg := range s.okf {
			msgs = append(msgs, fmsg(t1, t2))
		}
	} else {
		msgs = append(msgs, s.fail...)
		for _, fmsg := range s.failf {
			msgs = append(msgs, fmsg(err))
		}
	}

	printResult(start, err == nil, msgs)
	return err
}

func (s *Step[T1, T2]) Do1(f func() (T1, error)) (T1, error) {
	printTimerHeading(s.name)

	start := time.Now()
	t1, err := f()
	msgs := []string{}
	if err == nil {
		var t2 T2
		msgs = append(msgs, s.ok...)
		for _, fmsg := range s.okf {
			msgs = append(msgs, fmsg(t1, t2))
		}
	} else {
		msgs = append(msgs, s.fail...)
		for _, fmsg := range s.failf {
			msgs = append(msgs, fmsg(err))
		}
	}

	printResult(start, err == nil, msgs)
	return t1, err
}

func (s *Step[T1, T2]) Do2(f func() (T1, T2, error)) (T1, T2, error) {
	printTimerHeading(s.name)

	start := time.Now()
	t1, t2, err := f()
	msgs := []string{}
	if err == nil {
		msgs = append(msgs, s.ok...)
		for _, fmsg := range s.okf {
			msgs = append(msgs, fmsg(t1, t2))
		}
	} else {
		msgs = append(msgs, s.fail...)
		for _, fmsg := range s.failf {
			msgs = append(msgs, fmsg(err))
		}
	}

	printResult(start, err == nil, msgs)
	return t1, t2, err
}

func PrintHeading(title ...string) {
	fmt.Printf("• %s\n", FontBold+strings.Join(title, " ")+FontReset)
}

func printTimerHeading[S ~string](title S) {
	fmt.Printf("• %s...", title)
}

func printResult(start time.Time, ok bool, msgs []string) {
	if ok {
		fmt.Printf(FontBold+FontGreen+" OK"+FontReset+" (%s)\n", FontBold+time.Since(start).String()+FontReset)
		for _, msg := range msgs {
			fmt.Printf("   → %s\n", msg)
		}
	} else {
		fmt.Println(FontBold + FontRed + " FAIL" + FontReset)
		for _, msg := range msgs {
			fmt.Printf("   → %s\n", msg)
		}
	}
}
