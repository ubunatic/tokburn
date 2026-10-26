//go:build linux

package game

import (
	"os"
	"syscall"
	"unsafe"
)

func rawTerminal(in *os.File) (func(), bool, error) {
	var old syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, in.Fd(), syscall.TCGETS, uintptr(unsafe.Pointer(&old))); errno != 0 {
		if errno == syscall.ENOTTY {
			return func() {}, false, nil
		}
		return nil, false, errno
	}
	raw := old
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	set := func(state *syscall.Termios) syscall.Errno {
		_, _, err := syscall.Syscall(syscall.SYS_IOCTL, in.Fd(), syscall.TCSETS, uintptr(unsafe.Pointer(state)))
		return err
	}
	if err := set(&raw); err != 0 {
		return nil, false, err
	}
	return func() { set(&old) }, true, nil
}
