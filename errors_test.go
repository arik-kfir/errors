package errors_test

import (
	"bytes"
	"fmt"
	pkgerrors "github.com/pkg/errors"
	"testing"

	"github.com/secureworks/errors"
	"github.com/secureworks/errors/internal/testutils"
)

var (
	newMsg = "new err"
)

var (
	errorsTestFilM = `/errors_test\.go`
	errorTestFileM = func(line string) string { return fmt.Sprintf("^\\s*\t.+%s:%s$", errorsTestFilM, line) }
)

func TestErrorNew(t *testing.T) {
	t.Run("New not nil", func(t *testing.T) {
		err := errors.New("new1")
		testutils.AssertNotNil(t, err)
	})
	t.Run("New not nil", func(t *testing.T) {
		err := errors.New("new1: %s", "arg")
		testutils.AssertNotNil(t, err)
	})
	t.Run("Tags not added to message", func(t *testing.T) {
		err := errors.New("new1: %s", errors.Tag("synthetic"), "arg")
		testutils.AssertNotNil(t, err)
		testutils.AssertEqual(t, "new1: arg", err.Error())
		testutils.AssertEqual(t, 1, len(err.(errors.TagProvider).Tags()))
		testutils.AssertEqual(t, "synthetic", err.(errors.TagProvider).Tags()[0])
		testutils.AssertTrue(t, err.(errors.TagProvider).HasTag("synthetic"))
		testutils.AssertTrue(t, errors.HasTag(err, "synthetic"))
	})
	t.Run("Meta not added to message", func(t *testing.T) {
		err := errors.New("new1: %s", errors.Meta("user", "Joe"), "arg")
		testutils.AssertNotNil(t, err)
		testutils.AssertEqual(t, "new1: arg", err.Error())
		testutils.AssertEqual(t, 1, len(err.(errors.MetaProvider).MetaMap()))
		testutils.AssertEqual(t, "Joe", err.(errors.MetaProvider).MetaMap()["user"])
		testutils.AssertEqual(t, "Joe", err.(errors.MetaProvider).Meta("user"))
		testutils.AssertEqual(t, "Joe", errors.GetMeta(err, "user"))
	})
}

func TestErrorMessage(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		err := errors.New("new1")
		testutils.AssertEqual(t, "new1", err.Error())
	})
	t.Run("New", func(t *testing.T) {
		err := errors.New("new1: %s", "arg")
		testutils.AssertEqual(t, "new1: arg", err.Error())
	})
}

func TestErrorUnwrap(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		err := errors.New("new1")
		unwrapper, ok := err.(errors.Unwrapper)
		testutils.AssertTrue(t, ok)
		testutils.AssertNil(t, unwrapper.Unwrap())
	})
	t.Run("New without wrapped error", func(t *testing.T) {
		wrapper := errors.New("wrapper: %s", "root")
		unwrapper, ok := wrapper.(errors.Unwrapper)
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, wrapper.Error(), "wrapper: root")
		testutils.AssertNil(t, unwrapper.Unwrap())
	})
	t.Run("New with unprinted wrapped error", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper", root)
		unwrapper, ok := wrapper.(errors.Unwrapper)
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, wrapper.Error(), "wrapper")
		testutils.AssertNotNil(t, unwrapper.Unwrap())
		testutils.AssertEqual(t, root, unwrapper.Unwrap())
	})
	t.Run("New with printed wrapped error", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		unwrapper, ok := wrapper.(errors.Unwrapper)
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, wrapper.Error(), "wrapper: root")
		testutils.AssertNotNil(t, unwrapper.Unwrap())
		testutils.AssertEqual(t, root, unwrapper.Unwrap())
	})
}

func TestErrorStackTrace(t *testing.T) {
	t.Run("Standalone", func(t *testing.T) {
		err := errors.New("new1")
		st, ok := err.(errors.StackTracer)
		testutils.AssertTrue(t, ok)

		var frames errors.Frames
		for _, pc := range st.StackTrace() {
			frames = append(frames, errors.FrameFromPC(pc))
		}
		testutils.AssertLinesMatch(t, frames, "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorStackTrace.func1$",
				errorTestFileM("98"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
	t.Run("Wrapper", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		st, ok := wrapper.(errors.StackTracer)
		testutils.AssertTrue(t, ok)

		var frames errors.Frames
		for _, pc := range st.StackTrace() {
			frames = append(frames, errors.FrameFromPC(pc))
		}
		testutils.AssertLinesMatch(t, frames, "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorStackTrace.func2$",
				errorTestFileM("118"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
}

func TestErrorChainStackTrace(t *testing.T) {
	t.Run("Standalone", func(t *testing.T) {
		err := errors.New("new1")
		st, ok := err.(errors.ChainStackTracer)
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, 1, len(st.ChainStackTrace()))

		var frames errors.Frames
		for _, pc := range st.ChainStackTrace()[0] {
			frames = append(frames, errors.FrameFromPC(pc))
		}
		testutils.AssertLinesMatch(t, frames, "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorChainStackTrace.func1$",
				errorTestFileM("140"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
	t.Run("Wrapper", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		st, ok := wrapper.(errors.ChainStackTracer)
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, 2, len(st.ChainStackTrace()))

		var wrapperFrames errors.Frames
		for _, pc := range st.ChainStackTrace()[0] {
			wrapperFrames = append(wrapperFrames, errors.FrameFromPC(pc))
		}
		testutils.AssertLinesMatch(t, wrapperFrames, "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorChainStackTrace.func2$",
				errorTestFileM("161"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)

		var rootFrames errors.Frames
		for _, pc := range st.ChainStackTrace()[1] {
			rootFrames = append(rootFrames, errors.FrameFromPC(pc))
		}
		testutils.AssertLinesMatch(t, rootFrames, "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorChainStackTrace.func2$",
				errorTestFileM("160"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
}

func TestErrorFrames(t *testing.T) {
	t.Run("Standalone", func(t *testing.T) {
		err := errors.New("new1")
		fr, ok := err.(errors.Framer)
		testutils.AssertTrue(t, ok)
		testutils.AssertLinesMatch(t, fr.Frames(), "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorFrames.func1$",
				errorTestFileM("198"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
	t.Run("Wrapper", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		fr, ok := wrapper.(errors.Framer)
		testutils.AssertTrue(t, ok)
		testutils.AssertLinesMatch(t, fr.Frames(), "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorFrames.func2$",
				errorTestFileM("213"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
}

func TestErrorChainFrames(t *testing.T) {
	t.Run("Standalone", func(t *testing.T) {
		err := errors.New("new1")
		fr, ok := err.(errors.ChainFramer)
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, 1, len(fr.ChainFrames()))
		testutils.AssertLinesMatch(t, fr.ChainFrames()[0], "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorChainFrames.func1$",
				errorTestFileM("230"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
	t.Run("Wrapper", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		fr, ok := wrapper.(errors.ChainFramer)
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, 2, len(fr.ChainFrames()))
		testutils.AssertLinesMatch(t, fr.ChainFrames()[0], "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorChainFrames.func2$",
				errorTestFileM("246"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
		testutils.AssertLinesMatch(t, fr.ChainFrames()[1], "%+v",
			[]string{
				"",
				"^github\\.com/secureworks/errors_test.TestErrorChainFrames.func2$",
				errorTestFileM("245"),
				`^testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
}

func TestErrorFormat(t *testing.T) {
	cases := []struct {
		name   string
		format string
		error  error
		expect interface{}
	}{
		{"%s", "%s", errors.New("new1"), `new1`},
		{"%q", "%q", errors.New("new1"), `"new1"`},
		{"%v", "%v", errors.New("new1"), `new1`},
		{"%#v", "%#v", errors.New("new1"), `&errors.errorImpl{"new1" \%\!q\(\<nil\>\)}`},
		{"%d", "%d", errors.New("new1"), ``}, // empty
		{
			name:   "%+v of standalone error",
			format: "%+v",
			error:  errors.New("new1"),
			expect: []string{
				"new1",
				"^     github.com/secureworks/errors_test\\.TestErrorFormat$",
				errorTestFileM("286"),
				`^     testing\.tRunner$`,
				`^     .+/testing/testing.go:\d+$`,
			},
		},
		{
			name:   "%+v of wrapping error",
			format: "%+v",
			error:  errors.New("wrap2: %w", errors.New("wrap1: %w", errors.New("root"))),
			expect: []string{
				"wrap2: wrap1: root",
				"^     github.com/secureworks/errors_test.TestErrorFormat$",
				errorTestFileM("298"),
				`^     testing\.tRunner$`,
				`^     .+/testing/testing.go:\d+$`,
				"",
				"CAUSED BY: wrap1: root",
				"^     github.com/secureworks/errors_test.TestErrorFormat$",
				errorTestFileM("298"),
				`^     testing\.tRunner$`,
				`^     .+/testing/testing.go:\d+$`,
				"",
				"CAUSED BY: root",
				"^     github.com/secureworks/errors_test.TestErrorFormat$",
				errorTestFileM("298"),
				`^     testing\.tRunner$`,
				`^     .+/testing/testing.go:\d+$`,
			},
		},
		{
			name:   "%+v of chain with non-frames-error in the middle",
			format: "%+v",
			error:  errors.New("wrap2: %w", fmt.Errorf("wrap1: %w", errors.New("root"))),
			expect: []string{
				"wrap2: wrap1: root",
				"^     github.com/secureworks/errors_test.TestErrorFormat$",
				errorTestFileM("322"),
				`^     testing\.tRunner$`,
				`^     .+/testing/testing.go:\d+$`,
				"",
				"CAUSED BY: wrap1: root",
				"",
				"CAUSED BY: root",
				"^     github.com/secureworks/errors_test.TestErrorFormat$",
				errorTestFileM("322"),
				`^     testing\.tRunner$`,
				`^     .+/testing/testing.go:\d+$`,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			testutils.AssertLinesMatch(t, tt.error, tt.format, tt.expect)
		})
	}
}

func TestMask(t *testing.T) {
	t.Run("nil does nothing", func(t *testing.T) {
		testutils.AssertNil(t, errors.Mask(nil))
		testutils.AssertNil(t, errors.Mask((*errorType)(nil)))
	})
	t.Run("collapses wrapped errors, removing all information", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		wcst, wisChainStackTracer := wrapper.(errors.ChainStackTracer)
		testutils.AssertEqual(t, "wrapper: root", wrapper.Error())
		testutils.AssertTrue(t, errors.Is(wrapper, root))
		testutils.AssertTrue(t, wisChainStackTracer)
		testutils.AssertTrue(t, len(wcst.ChainStackTrace()) == 2)

		masked := errors.Mask(wrapper)
		mcst, misChainStackTracer := masked.(errors.ChainStackTracer)
		testutils.AssertEqual(t, "wrapper: root", masked.Error())
		testutils.AssertFalse(t, errors.Is(masked, root))
		testutils.AssertTrue(t, misChainStackTracer)
		testutils.AssertTrue(t, len(mcst.ChainStackTrace()) == 1) // <- now it's length is 1 rather than 2 (single stacktrace)
	})
}

func TestOpaque(t *testing.T) {
	t.Run("nil does nothing", func(t *testing.T) {
		testutils.AssertNil(t, errors.Opaque(nil))
		testutils.AssertNil(t, errors.Opaque((*errorType)(nil)))
	})
	t.Run("collapses wrapped errors, but retains frames", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		wcst, wisChainStackTracer := wrapper.(errors.ChainStackTracer)
		testutils.AssertEqual(t, "wrapper: root", wrapper.Error())
		testutils.AssertTrue(t, errors.Is(wrapper, root))
		testutils.AssertTrue(t, wisChainStackTracer)
		testutils.AssertTrue(t, len(wcst.ChainStackTrace()) == 2)

		opaque := errors.Opaque(wrapper)
		_, oisChainStackTracer := opaque.(errors.ChainStackTracer)
		testutils.AssertEqual(t, "wrapper: root", opaque.Error())
		testutils.AssertFalse(t, errors.Is(opaque, wrapper))
		testutils.AssertTrue(t, oisChainStackTracer)
	})
}

func TestPrintStackChain(t *testing.T) {
	t.Run("nil writer panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				testutils.AssertEqual(t, "nil writer", r)
			} else {
				t.Error("expected panic")
			}
		}()
		errors.PrintStackChain(nil, errors.New("oops"))
	})
	t.Run("prints stack trace of standalone error", func(t *testing.T) {
		err := errors.New("new1")
		buf := bytes.Buffer{}
		errors.PrintStackChain(&buf, err)
		testutils.AssertLinesMatch(t, buf.String(), "%s",
			[]string{
				"new1",
				"^     github\\.com/secureworks/errors_test\\.TestPrintStackChain.func2$",
				errorTestFileM("404"),
				`^     testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
	t.Run("prints stack trace of wrapping error", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		buf := bytes.Buffer{}
		errors.PrintStackChain(&buf, wrapper)
		testutils.AssertLinesMatch(t, buf.String(), "%s",
			[]string{
				"wrapper: root",
				"^     github\\.com/secureworks/errors_test\\.TestPrintStackChain.func3$",
				errorTestFileM("419"),
				`^     testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
				"",
				"CAUSED BY: root",
				"^     github\\.com/secureworks/errors_test\\.TestPrintStackChain.func3$",
				errorTestFileM("418"),
				`^     testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
	t.Run("prints stack trace of no-stack error wrapping stacked error", func(t *testing.T) {
		root := errors.New("root")
		wrapper := errors.New("wrapper: %w", root)
		buf := bytes.Buffer{}
		errors.PrintStackChain(&buf, wrapper)
		testutils.AssertLinesMatch(t, buf.String(), "%s",
			[]string{
				"wrapper: root",
				"^     github\\.com/secureworks/errors_test\\.TestPrintStackChain.func4$",
				errorTestFileM("440"),
				`^     testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
				"",
				"CAUSED BY: root",
				"^     github\\.com/secureworks/errors_test\\.TestPrintStackChain.func4$",
				errorTestFileM("439"),
				`^     testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
	t.Run("prints stack trace of standalone pkg/errors error", func(t *testing.T) {
		err := pkgerrors.WithStack(fmt.Errorf("root"))
		buf := bytes.Buffer{}
		errors.PrintStackChain(&buf, err)
		testutils.AssertLinesMatch(t, buf.String(), "%s",
			[]string{
				"root",
				"^     github\\.com/secureworks/errors_test\\.TestPrintStackChain.func5$",
				errorTestFileM("460"),
				`^     testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
				`^     runtime.goexit$`,
				`^.+/runtime/.+:\d+$`,
				"",
				"CAUSED BY: root",
			},
		)
	})
	t.Run("prints stack trace of no-stack error wrapping stacked error", func(t *testing.T) {
		root := errors.New("root")
		wrapper := pkgerrors.Wrapf(root, "wrapper")
		buf := bytes.Buffer{}
		errors.PrintStackChain(&buf, wrapper)
		testutils.AssertLinesMatch(t, buf.String(), "%s",
			[]string{
				"wrapper: root",
				"^     github\\.com/secureworks/errors_test\\.TestPrintStackChain.func6$",
				errorTestFileM("479"),
				`^     testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
				`^     runtime.goexit$`,
				`^.+/runtime/.+:\d+$`,
				"",
				"CAUSED BY: wrapper: root",
				"",
				"CAUSED BY: root",
				"^     github\\.com/secureworks/errors_test\\.TestPrintStackChain.func6$",
				errorTestFileM("478"),
				`^     testing\.tRunner$`,
				`^.+/testing/testing.go:\d+$`,
			},
		)
	})
}
