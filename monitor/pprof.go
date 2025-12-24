package monitor

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/songquanpeng/one-api/common/env"
	"github.com/songquanpeng/one-api/common/logger"
)

// InitPprof initializes pprof HTTP server for performance profiling
func InitPprof() {
	if !env.Bool("PPROF_ENABLED", false) {
		return
	}

	port := env.String("PPROF_PORT", "6060")
	addr := fmt.Sprintf(":%s", port)

	logger.SysLog(fmt.Sprintf("pprof server starting on http://localhost:%s/debug/pprof/", port))
	logger.SysLog("⚠️  WARNING: pprof exposes runtime information - use only in trusted environments")
	logger.SysLog("Available endpoints:")
	logger.SysLog("  - CPU profile: http://localhost:" + port + "/debug/pprof/profile?seconds=30")
	logger.SysLog("  - Heap profile: http://localhost:" + port + "/debug/pprof/heap")
	logger.SysLog("  - Goroutines: http://localhost:" + port + "/debug/pprof/goroutine?debug=2")
	logger.SysLog("  - Block profile: http://localhost:" + port + "/debug/pprof/block?debug=2")
	logger.SysLog("  - Mutex profile: http://localhost:" + port + "/debug/pprof/mutex?debug=2")

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.SysError(fmt.Sprintf("pprof server error: %v", err))
			os.Exit(1)
		}
	}()
}
