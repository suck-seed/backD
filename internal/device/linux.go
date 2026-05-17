package device

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// isWritable concluded if the mount drive is writeable or not
func isWritable(path string) bool {

	ok, err := mountOptionsWritable(path)
	if err != nil || !ok {
		return false
	}

	if !accessWritable(path) {
		return false
	}

	//? final confirmation for network or critical ops:
	testFile := filepath.Join(path, ".backd_write_test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)

	return true
}

func mountOptionsWritable(mount string) (bool, error) {

	// Reading /proc/mount
	// Especially line[2] -> FileSystemType
	// line[0] -> /dev/...
	// line[1] -> mount
	// line[4] -> rw/ro

	f, err := os.Open("/proc/mounts")
	if err != nil {
		return false, err
	}
	defer f.Close()

	// line by line scanner, using bufio.NewScanner
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		// Fetch the line into fields
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}

		// mount in fields contains \040 for space
		// fetch fields[1] (mount) without \040
		mountPoint := strings.ReplaceAll(fields[1], `\040`, " ")
		if mountPoint != mount {
			continue
		}

		for o := range strings.SplitSeq(fields[3], ",") {
			if o == "ro" {
				return false, nil
			}
			if o == "rw" {
				return true, nil
			}
		}

		return false, nil

	}

	if err := scanner.Err(); err != nil {
		return false, err
	}
	return false, fmt.Errorf("Mount point not found")
}

// fsTypeForMount returns the filesystem type such as NTFS, btfs, exFAT, FAT32 and so on
func fsTypeForMount(mount string) (string, error) {

	// Reading /proc/mount
	// Especially line[2] -> FileSystemType
	// line[0] -> /dev/...
	// line[1] -> mount

	f, err := os.Open("/proc/mounts")
	if err != nil {
		return "", err
	}
	defer f.Close()

	// line by line scanner, using bufio.NewScanner
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		// Fetch the line into fields
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		// mount in fields contains \040 for space
		// fetch fields[1] (mount) without \040
		mountPoint := strings.ReplaceAll(fields[1], `\040`, " ")
		if mountPoint == mount {
			return fields[2], nil
		}

	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("Mount point not found")
}

func accessWritable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

// bytesFromStatfs returns total and available bytes for path.
func bytesFromStatfs(path string) (total uint64, avail uint64, free uint64, err error) {
	var st unix.Statfs_t

	err = unix.Statfs(path, &st)
	if err != nil {
		return 0, 0, 0, err
	}

	// Bsize   int64
	// Blocks  uint64
	// Bfree   uint64
	// Bavail  uint64
	bSize := uint64(st.Bsize)
	total = uint64(st.Blocks) * bSize
	avail = uint64(st.Bavail) * bSize
	free = uint64(st.Bfree) * bSize

	return total, avail, free, nil

}
