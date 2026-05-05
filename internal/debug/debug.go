//go:build debug

package debug

import "log"

func DebugLog(format string, args ...any) {
	if verbose { // verbose is your -v flag
		log.Printf("[DEBUG] "+format, args...)
	}
}
