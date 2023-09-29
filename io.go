package main

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
		n = Read(fd, buf[read:])
		if n == 0 {
			return read
		} else if n < 0 {
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
		n = Write(fd, buf[written:])
		if n == 0 {
			return written
		} else if n < 0 {
			if -n != EINTR {
				return n
			}
			continue
		}
		written += n
	}

	return int64(len(buf))
}

func ReadEntireFile(fd int32) ([]byte, error) {
	var flen int64
	if flen = Lseek(fd, 0, SEEK_END); flen < 0 {
		return nil, NewError("Failed to get file length: ", int(flen))
	}
	data := make([]byte, flen)
	if ret := Lseek(fd, 0, SEEK_SET); ret < 0 {
		return nil, NewError("Failed to seek to the beginning of the file: ", int(flen))
	}
	if n := ReadFull(fd, data); n < 0 {
		return nil, NewError("Failed to read entire file: ", int(n))
	}

	return data, nil
}
