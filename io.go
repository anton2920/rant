package main

/* See <sys/stat.h>. */
type Stat struct {
	Dev       uint   /* inode's device */
	Ino       uint   /* inode's number */
	Nlink     uint64 /* number of hard links */
	Mode      uint16 /* inode protection mode */
	_         int16
	Uid       uint32 /* user ID of the file's owner */
	Gid       uint32 /* group ID of the file's group */
	_         int32
	Rdev      uint64   /* device type */
	Atime     Timespec /* time of last access */
	Mtime     Timespec /* time of last data modification */
	Ctime     Timespec /* time of last file status change */
	Birthtime Timespec /* time of file creation */
	Size      int      /* file size, in bytes */
	Blocks    int      /* blocks allocated for file */
	Blksize   int32    /* optimal blocksize for I/O */
	Flags     uint32   /* user defined flags for file */
	Gen       uint64   /* file generation number */
	_         [10]int
}

const (
	/* From <fcntl.h>. */
	O_RDONLY = 0

	SEEK_SET = 0
	SEEK_END = 2

	PATH_MAX = 1024
)

func ReadFull(fd int32, buf []byte) int64 {
	var read, n int64
	for read < int64(len(buf)) {
		if n = Read(fd, buf[read:]); n < 0 {
			if -n != EINTR {
				return n
			}
			continue
		}
		read += n
	}

	return int64(len(buf))
}

func WriteFull(fd int32, buf []byte) int64 {
	var written, n int64
	for written < int64(len(buf)) {
		if n = Write(fd, buf[written:]); n < 0 {
			if -n != EINTR {
				return n
			}
			continue
		}
		written += n
	}

	return int64(len(buf))
}

func ReadEntireFile(fd int32) []byte {
	var flen int64
	if flen = Lseek(fd, 0, SEEK_END); flen < 0 {
		Fatal("Failed to get file length: ", int(flen))
	}
	data := make([]byte, flen)
	if ret := Lseek(fd, 0, SEEK_SET); ret < 0 {
		Fatal("Failed to seek to the beginning of the file: ", int(flen))
	}
	if n := ReadFull(fd, data); n < 0 {
		Fatal("Failed to read entire file: ", int(n))
	}

	return data
}
