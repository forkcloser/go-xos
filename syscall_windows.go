//go:build windows

package xos

import (
	"syscall"
	"unsafe"
)

// What xos needs of the Windows API that Go's syscall package does not have.
// Go's own os takes these from internal/syscall/windows, which nothing
// outside the standard library can import. The values are Windows's, from
// winnt.h, winbase.h and winerror.h.

// kernel32 is loaded from the system directory alone: syscall lists
// kernel32.dll among the system libraries itself, under that exact name.
//
// Three calls are bound here. SetFileInformationByHandle is not in syscall.
// SetEndOfFile and GetFileInformationByHandle are, and are bound here all the
// same for how their result is read (see done).
//
//nolint:gochecknoglobals // a lazily bound system library
var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procSetFileInformationByHandle = kernel32.NewProc("SetFileInformationByHandle")
	procSetEndOfFile               = kernel32.NewProc("SetEndOfFile")
	procGetFileInformationByHandle = kernel32.NewProc("GetFileInformationByHandle")
)

const (
	// accessDelete is the DELETE access right.
	accessDelete = 0x00010000

	// errInvalidParameter is ERROR_INVALID_PARAMETER.
	errInvalidParameter syscall.Errno = 87

	// The FILE_FLAG_* of CreateFile that syscall has no name for.
	fileFlagWriteThrough   = 0x80000000
	fileFlagNoBuffering    = 0x20000000
	fileFlagRandomAccess   = 0x10000000
	fileFlagSequentialScan = 0x08000000
	fileFlagDeleteOnClose  = 0x04000000
	fileFlagPOSIXSemantics = 0x01000000
	fileFlagSessionAware   = 0x00800000
	fileFlagOpenNoRecall   = 0x00100000

	// fileRenameInfoEx is the FileRenameInfoEx class of
	// SetFileInformationByHandle, and the two flags its FILE_RENAME_INFO
	// takes here.
	fileRenameInfoEx          = 22
	fileRenameReplaceIfExists = 0x1
	fileRenamePOSIXSemantics  = 0x2
)

// done is the error of a call that returns a BOOL: none when the BOOL is
// true. A BOOL is 32 bits, and on 64-bit Windows the upper half of the
// register it comes back in is undefined: a kernel32 that something hooks
// may leave anything there, and a result read whole would then take a
// failure for a success. syscall's own wrappers read it whole in Go 1.26
// and 1.27.
func done(result uintptr, lastErr syscall.Errno) error {
	if uint32(result) != 0 { // #nosec G115 -- the low half is the BOOL
		return nil
	}

	// A failure that left no error code is still a failure.
	if lastErr == 0 {
		return syscall.EINVAL
	}

	return lastErr
}

// setFileInformationByHandle is SetFileInformationByHandle.
func setFileInformationByHandle(handle syscall.Handle, class uint32, info unsafe.Pointer, size uint32) error {
	result, _, lastErr := syscall.SyscallN(
		procSetFileInformationByHandle.Addr(),
		uintptr(handle),
		uintptr(class),
		uintptr(info),
		uintptr(size),
	)

	return done(result, lastErr)
}

// setEndOfFile is SetEndOfFile.
func setEndOfFile(handle syscall.Handle) error {
	result, _, lastErr := syscall.SyscallN(procSetEndOfFile.Addr(), uintptr(handle))

	return done(result, lastErr)
}

// getFileInformationByHandle is GetFileInformationByHandle.
func getFileInformationByHandle(handle syscall.Handle, about *syscall.ByHandleFileInformation) error {
	result, _, lastErr := syscall.SyscallN(
		procGetFileInformationByHandle.Addr(),
		uintptr(handle),
		uintptr(unsafe.Pointer(about)), // #nosec G103 -- the struct is what the call fills
	)

	return done(result, lastErr)
}
