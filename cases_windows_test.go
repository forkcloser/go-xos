//go:build windows

package xos_test

import (
	"os"
	"testing"

	"github.com/forkcloser/go-xos"
)

// TestFileFlags: the FILE_FLAG_* bits of an open flag reach CreateFile, and
// a bit there that is no such flag is refused, as with os.OpenFile.
func TestFileFlags(t *testing.T) {
	t.Parallel()

	// The FILE_FLAG_* values, as Windows documents them for CreateFile.
	for _, flag := range []int{
		0x08000000, // FILE_FLAG_SEQUENTIAL_SCAN
		0x10000000, // FILE_FLAG_RANDOM_ACCESS
		0x40000000, // FILE_FLAG_OVERLAPPED
		0x80000000, // FILE_FLAG_WRITE_THROUGH
		0x04000000, // FILE_FLAG_DELETE_ON_CLOSE
		0x00400000, // within the mask, no FILE_FLAG
	} {
		ours, theirs := pair(t, func(t *testing.T, path string) {
			t.Helper()

			if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
				t.Fatal(err)
			}
		})

		got := openAndUse(xos.OpenFile, ours, os.O_RDWR|flag, 0o644)
		want := openAndUse(os.OpenFile, theirs, os.O_RDWR|flag, 0o644)

		if got != want {
			t.Fatalf("flag %#x: xos %q, os %q", flag, got, want)
		}

		if got, want := onDisk(ours), onDisk(theirs); got != want {
			t.Fatalf("flag %#x: xos left %q, os %q", flag, got, want)
		}
	}
}
