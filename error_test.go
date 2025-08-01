package stackerr

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	e := New("fooey")
	require.Error(t, e)
	require.Equal(t, "fooey", e.Error())
}

func TestNewf(t *testing.T) {
	e := Newf("fooey %d", 2)
	require.Error(t, e)
	require.Equal(t, "fooey 2", e.Error())
}

func TestWrap(t *testing.T) {
	DefaultPackageName = "stackerr"
	defer func() {
		DefaultPackageName = ""
	}()
	e := Wrap(errors.New("cause"), "fooey")
	require.Error(t, e)
	require.Equal(t, "fooey", e.Error())
	e2 := errors.Unwrap(e)
	require.Error(t, e2)
	require.Equal(t, "cause", e2.Error())

	si := e.StackInfo()
	require.Len(t, si, 1)
	require.Equal(t, 28, si[0].Line)

	require.NoError(t, Wrap(nil, "fooey"))
}

func TestError_Unwrap(t *testing.T) {
	e := New("fooey")
	require.Error(t, e)
	e2 := e.Unwrap()
	require.NoError(t, e2)
	require.NoError(t, errors.Unwrap(e))
	e = e.WithCause(errors.New("cause"))
	e2 = e.Unwrap()
	require.Error(t, e2)
	require.Error(t, errors.Unwrap(e))
}

func TestError_Cause(t *testing.T) {
	e := New("fooey")
	require.Error(t, e)
	e2 := e.Cause()
	require.NoError(t, e2)
	e = e.WithCause(errors.New("cause"))
	e2 = e.Cause()
	require.Error(t, e2)
}

func TestError_StackInfo(t *testing.T) {
	t.Run("with DefaultPackageName", func(t *testing.T) {
		DefaultPackageName = "stackerr"
		defer func() {
			DefaultPackageName = ""
		}()
		e := New("fooey")
		require.Error(t, e)
		si := e.StackInfo()
		require.Len(t, si, 1)
		require.True(t, strings.HasPrefix(si[0].Function, "github.com/go-andiamo/stackerr"))
		require.Contains(t, si[0].Function, "TestError_StackInfo")
		require.Equal(t, 70, si[0].Line)
	})
	t.Run("with DefaultPackageFilter", func(t *testing.T) {
		DefaultPackageFilter = &testPackageFilter{}
		defer func() {
			DefaultPackageFilter = nil
		}()
		e := New("fooey")
		require.Error(t, e)
		si := e.StackInfo()
		require.Len(t, si, 1)
		require.True(t, strings.HasPrefix(si[0].Function, "github.com/go-andiamo/stackerr"))
		require.Contains(t, si[0].Function, "TestError_StackInfo")
		require.Equal(t, 83, si[0].Line)
	})
	t.Run("with SetDefaultPackageFilter", func(t *testing.T) {
		SetDefaultPackageFilter("github.com/go-andiamo/stackerr")
		defer func() {
			DefaultPackageFilter = nil
		}()
		e := New("fooey")
		require.Error(t, e)
		si := e.StackInfo()
		require.Len(t, si, 1)
		require.True(t, strings.HasPrefix(si[0].Function, "github.com/go-andiamo/stackerr"))
		require.Contains(t, si[0].Function, "TestError_StackInfo")
		require.Equal(t, 96, si[0].Line)
	})
}

type testPackageFilter struct{}

var _ PackageFilter = (*testPackageFilter)(nil)

func (pf *testPackageFilter) Include(packageName string) bool {
	return strings.Contains(packageName, "stackerr")
}

func TestError_Format(t *testing.T) {
	t.Run("v", func(t *testing.T) {
		e := New("fooey")
		require.Error(t, e)
		require.Equal(t, "fooey", fmt.Sprintf("%v", e))
	})
	t.Run("v with cause", func(t *testing.T) {
		e := New("fooey").WithCause(errors.New("cause"))
		require.Error(t, e)
		require.Equal(t, "fooey: cause", fmt.Sprintf("%v", e))
	})
	t.Run("+v", func(t *testing.T) {
		DefaultPackageName = "stackerr"
		defer func() {
			DefaultPackageName = ""
		}()
		e := New("fooey")
		require.Error(t, e)
		out := fmt.Sprintf("%+v", e)
		lines := strings.Split(out, "\n")
		require.Len(t, lines, 3)
		require.Equal(t, "fooey", lines[0])
		require.Equal(t, "Stack:", lines[1])
		require.True(t, strings.HasPrefix(lines[2], "\tgithub.com/go-andiamo/stackerr."))
		require.Contains(t, lines[2], ".TestError_Format.")
		require.True(t, strings.HasSuffix(lines[2], ":130"))
	})
	t.Run("+v with cause", func(t *testing.T) {
		DefaultPackageName = "stackerr"
		defer func() {
			DefaultPackageName = ""
		}()
		e := New("fooey").WithCause(errors.New("cause"))
		require.Error(t, e)
		out := fmt.Sprintf("%+v", e)
		lines := strings.Split(out, "\n")
		require.Len(t, lines, 3)
		require.Equal(t, "fooey: cause", lines[0])
		require.Equal(t, "Stack:", lines[1])
		require.True(t, strings.HasPrefix(lines[2], "\tgithub.com/go-andiamo/stackerr."))
		require.Contains(t, lines[2], ".TestError_Format.")
		require.True(t, strings.HasSuffix(lines[2], ":146"))
	})
	t.Run("s", func(t *testing.T) {
		e := New("fooey")
		require.Error(t, e)
		require.Equal(t, "fooey", fmt.Sprintf("%s", e))
	})
	t.Run("q", func(t *testing.T) {
		e := New("fooey")
		require.Error(t, e)
		require.Equal(t, `"fooey"`, fmt.Sprintf("%q", e))
	})
	t.Run("unknown", func(t *testing.T) {
		e := New("fooey")
		require.Error(t, e)
		require.Equal(t, "%!d(stackerr.err)", fmt.Sprintf("%d", e))
	})
}

func TestPackageFromFunction(t *testing.T) {
	full, short := packageFromFunction("github.com/go-andiamo/stackerr.TestSomething.func2")
	require.Equal(t, "github.com/go-andiamo/stackerr", full)
	require.Equal(t, "stackerr", short)

	full, short = packageFromFunction("stackerr.TestSomething.func2")
	require.Equal(t, "stackerr", full)
	require.Equal(t, "stackerr", short)
}
