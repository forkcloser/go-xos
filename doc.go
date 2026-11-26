// Package xos is the standard library's os, with one difference on Windows:
// every file it opens or creates carries FILE_SHARE_DELETE, so the file can
// be renamed or removed while it is open, as on every other platform, and
// [Rename] replaces a file held open that way where the volume allows it.
// Everything else is os, call for call: the same names, the same results,
// the same errors.
//
// # Where the Windows implementation parts from os
//
// On Windows [OpenFile] is a port of syscall.Open with the share mode widened,
// and the file it returns is made with os.NewFile, os having no way to open
// with another share mode. It parts from os.OpenFile in four places:
//
//   - Long paths. os puts the \\?\ prefix on a long path when the process
//     cannot opt into long paths, which is Windows before 10.0.15063. xos adds
//     no prefix.
//   - The directory bit. os.OpenFile refuses what is no directory when the
//     flag carries 0x4000, the unexported bit os.ReadDir opens with. OpenFile
//     ignores the bit; [ReadDir] makes the same check after its open.
//   - O_APPEND. os.NewFile does not record that a file was opened for
//     appending, so OpenFile sets the unexported mark of os.File itself, and
//     fails when that field is not as it expects.
//   - Overlapped I/O. Given a handle through os.NewFile, Go asks the system
//     whether it is overlapped, one more call for each open, and may leave
//     its completion notification modes as they are (Go 1.27.4 does, Go
//     1.26.0 does not), where os.OpenFile sets
//     FILE_SKIP_COMPLETION_PORT_ON_SUCCESS: a file opened with
//     O_FILE_FLAG_OVERLAPPED then waits on the completion port for an
//     operation that completed at once.
package xos
