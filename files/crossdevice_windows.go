package files

import "syscall"

// errCrossDevice is what rename returns for a move across volumes:
// ERROR_NOT_SAME_DEVICE, as MoveFileEx runs without MOVEFILE_COPY_ALLOWED.
// syscall.EXDEV is a placeholder on Windows that no call returns.
var errCrossDevice error = syscall.Errno(17)
