package handlers

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func TrimFilePath(path string) string {
	if path == emptyString {
		return emptyString
	}

	parts := strings.Split(filepath.ToSlash(path), pathSeparator)
	if len(parts) >= defaultMaxBackups {
		return strings.Join(parts[len(parts)-defaultMaxBackups:], pathSeparator)
	}

	return parts[0]
}

func TrimFuncName(fn string) string {
	if fn == emptyString {
		return emptyString
	}

	if i := strings.LastIndexByte(fn, pathSlash); i != -1 {
		fn = fn[i+1:]
	}

	if i := strings.LastIndexByte(fn, funcSeparator); i != -1 {
		fn = fn[i+1:]
	}

	return fn
}

func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := f.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
