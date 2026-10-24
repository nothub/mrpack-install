//go:build !windows

package files

import "syscall"

// errCrossDevice is what rename returns for a move across filesystems.
var errCrossDevice error = syscall.EXDEV
