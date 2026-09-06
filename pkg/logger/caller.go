package logger

import (
	"runtime"
	"sync"
)

const (
	maxStackDepth      = 32
	defaultCacheSize   = 1024
	maxCacheSize       = 10000
	defaultCallersSkip = 4
)

type callerCache struct {
	mu    sync.RWMutex
	cache map[uintptr]callerInfo
}

type callerInfo struct {
	pc       uintptr
	file     string
	line     int
	funcName string
}

var (
	globalCache = &callerCache{
		cache: make(map[uintptr]callerInfo, defaultCacheSize),
	}
)

func getCallerPCFast() (uintptr, string, int, string) {
	var pcs [maxStackDepth]uintptr

	num := runtime.Callers(defaultCallersSkip, pcs[:])
	if num == 0 {
		return 0, emptyString, 0, emptyString
	}

	callsite := pcs[0]
	if info, ok := globalCache.get(callsite); ok {
		return info.pc, info.file, info.line, info.funcName
	}

	frames := runtime.CallersFrames(pcs[:num])

	for {
		frame, _ := frames.Next()

		globalCache.set(callsite, callerInfo{
			pc:       frame.PC,
			file:     frame.File,
			line:     frame.Line,
			funcName: frame.Function,
		})

		return frame.PC, frame.File, frame.Line, frame.Function
	}
}

func (c *callerCache) get(key uintptr) (callerInfo, bool) {
	c.mu.RLock()
	info, ok := c.cache[key]
	c.mu.RUnlock()

	return info, ok
}

func (c *callerCache) set(key uintptr, info callerInfo) {
	c.mu.Lock()
	if len(c.cache) > maxCacheSize {
		c.cache = make(map[uintptr]callerInfo, defaultCacheSize)
	}

	c.cache[key] = info
	c.mu.Unlock()
}
