//go:build windows

package xos

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Rename renames (moves) oldpath to newpath, as os.Rename does, and on a volume
// that has the POSIX-semantics rename (NTFS) replaces a newpath another handle
// holds open with FILE_SHARE_DELETE, as a rename does on Unix. A volume without
// it (FAT, exFAT, some shares) gets os.Rename (MoveFileEx), which refuses a
// held newpath: the property is the volume's. A directory is moved as os.Rename
// moves one.
//
//nolint:wrapcheck // Thin wrapper matching os.Rename signature.
func Rename(oldpath, newpath string) error {
	if posixRename(oldpath, newpath) {
		return nil
	}

	return os.Rename(oldpath, newpath)
}

// fileRenameInfo is FILE_RENAME_INFO with the Flags member of its union, as
// FileRenameInfoEx reads it.
type fileRenameInfo struct {
	Flags          uint32
	RootDirectory  syscall.Handle
	FileNameLength uint32
	FileName       [syscall.MAX_LONG_PATH]uint16
}

// posixRename asks the filesystem to take newpath's name over at once, a handle
// still open on the old newpath keeping the old file until it closes, and
// reports whether it did. NTFS does, every time a rename can succeed at all;
// FAT, exFAT and some shares refuse the request itself, every time, and
// os.Rename then does there what it does.
func posixRename(oldpath, newpath string) bool {
	target, err := filepath.Abs(newpath)
	if err != nil {
		return false
	}

	name, err := syscall.UTF16FromString(target)
	if err != nil || len(name) > syscall.MAX_LONG_PATH {
		return false
	}

	source, err := syscall.UTF16PtrFromString(oldpath)
	if err != nil {
		return false
	}

	handle, err := syscall.CreateFile(
		source,
		accessDelete|syscall.SYNCHRONIZE,
		shareMode,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return false
	}

	defer func() { _ = syscall.CloseHandle(handle) }()

	// A directory goes to os.Rename: the POSIX rename would replace an empty
	// directory in its way, as Unix does, where MoveFileEx refuses, and xos
	// parts from os over held files alone.
	var about syscall.ByHandleFileInformation
	if getFileInformationByHandle(handle, &about) != nil ||
		about.FileAttributes&syscall.FILE_ATTRIBUTE_DIRECTORY != 0 {
		return false
	}

	info := fileRenameInfo{Flags: fileRenameReplaceIfExists | fileRenamePOSIXSemantics}
	copy(info.FileName[:], name)
	info.FileNameLength = uint32((len(name) - 1) * 2) // #nosec G115 -- bounded by MAX_LONG_PATH above

	return setFileInformationByHandle(
		handle,
		fileRenameInfoEx,
		unsafe.Pointer(&info), // #nosec G103 -- the struct is what the call reads
		uint32(unsafe.Sizeof(info)),
	) == nil
}
