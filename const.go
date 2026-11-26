package xos

import "os"

const (
	minReadBuf = 512

	// defaultPerm is the standard Unix permission for os.Create (before umask).
	defaultPerm os.FileMode = 0o666

	// tempFilePerm and tempDirPerm are os.CreateTemp's and os.MkdirTemp's.
	tempFilePerm os.FileMode = 0o600
	tempDirPerm  os.FileMode = 0o700
)
