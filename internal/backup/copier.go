package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type CopyResult int

const (
	CopyResultCopied  CopyResult = iota // new file
	CopyResultUpdated                   // existing file overwritten
	CopyResultSkipped                   // destination indentical - nothing to do
	CopyResultFailed                    // error
)

// SafeCopy copies src to dst using a write-to-temp-then-rename strategy so
// that the destination is never left in a partially written state.
//
// It always creates any missing parent directories for dst.
// It preserves the source file's permission bits and modification time.
// It never follows symlinks — if src is a symlink the call returns an error.
//
// The caller is responsible for deciding whether the copy is needed; SafeCopy
// always performs the copy unconditionally (it does not re-check metadata).
//
// On success it returns the number of bytes written and CopyResultCopied or
// CopyResultUpdated (the distinction must be supplied by the caller via the
// existed parameter).
func SafeCopy(src, dst string, existed bool) (int64, CopyResult, error) {

	// Check for symlnlik
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return 0,
			CopyResultFailed,
			fmt.Errorf("Source Stat %s : %w", src, err)
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return 0,
			CopyResultFailed,
			fmt.Errorf("Source %s is a symlink - skipped", src)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return 0,
			CopyResultFailed,
			fmt.Errorf("create destination directory for %s: %w", src, err)
	}

	// Open source
	in, err := os.Open(src)
	if err != nil {
		return 0,
			CopyResultFailed,
			fmt.Errorf("open source %s: %w", src, err)
	}

	defer in.Close()

	// Write to a temp file in the same directory so the rename is atomic
	// on most file systems	(same file mount point)

	tmp := dst + ".backd_tmp"

	// Open the tmp file
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return 0,
			CopyResultFailed,
			fmt.Errorf("create temp file %s: %w", tmp, err)
	}

	// Copy from in -> out
	written, copyErr := io.Copy(out, in)
	closeErr := out.Close()

	if copyErr != nil {
		os.Remove(tmp)
		return 0,
			CopyResultFailed,
			fmt.Errorf("copy data %s -> %s: %w", src, dst, err)
	}

	if closeErr != nil {
		os.Remove(tmp)
		return 0, CopyResultFailed, fmt.Errorf("flush %s: %w", tmp, closeErr)

	}

	// Atomic rename
	if err := os.Rename(tmp, dst); err != nil {
		os.Remove(tmp)
		return 0, CopyResultFailed, fmt.Errorf("rename %s -> %s: %w", tmp, dst, err)

	}

	// preserver modification time
	os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())

	result := CopyResultCopied
	if existed {
		result = CopyResultUpdated
	}

	return written, result, nil

}

// MirrorSrcToDst builds mirror destination path for a source file under the backD root on the external drive
func MirrorSrcToDst(src, deviceMount string) string {

	// normalize
	relative := filepath.Clean(src)

	// Remove the leading /
	// if len(relative) > 0 && relative[0] == '/' {
	// 	relative = relative[1:]
	// }
	// if len(relative) > 0 && relative[0] == os.PathSeparator {
	// 	relative = relative[1:]
	// }

	if len(relative) > 0 && os.IsPathSeparator(relative[0]) {
		relative = relative[1:]
	}

	// will return filepath in this format
	// /run/media/user/Samsung_T7/backD/home/user/Documents/report.pdf	// if len(relative) > 0 && relative[0] == os.PathSeparator {
	// 	relative = relative[1:]
	// }

	return filepath.Join(deviceMount, BackDFolder, relative)

}
