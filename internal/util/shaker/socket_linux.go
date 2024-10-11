//go:build linux
// +build linux

package shaker

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"golang.org/x/sys/unix"
)

const maxEpollEvents = 32

func createSocketZeroLinger(family int, zeroLinger bool) (fd int, err error) {
	fd, err = _createNonBlockingSocket(family)
	if err == nil {
		if zeroLinger {
			err = _setZeroLinger(fd)
		}
	}
	return
}

func _createNonBlockingSocket(family int) (int, error) {
	fd, err := _createSocket(family)
	if err != nil {
		return 0, err
	}
	err = _setSockOpts(fd)
	if err != nil {
		unix.Close(fd)
	}
	return fd, err
}

func _createSocket(family int) (int, error) {
	fd, err := unix.Socket(family, unix.SOCK_STREAM, 0)
	unix.CloseOnExec(fd)
	return fd, err
}

func _setSockOpts(fd int) error {
	err := unix.SetNonblock(fd, true)
	if err != nil {
		return err
	}
	return unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_QUICKACK, 0)
}

var zeroLinger = unix.Linger{Onoff: 1, Linger: 0}

func _setZeroLinger(fd int) error {
	return unix.SetsockoptLinger(fd, unix.SOL_SOCKET, unix.SO_LINGER, &zeroLinger)
}

func createPoller() (fd int, err error) {
	fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		err = os.NewSyscallError("epoll_create1", err)
	}
	return fd, err
}

func registerEvents(pollerFd int, fd int) error {
	var event unix.EpollEvent
	event.Events = unix.EPOLLOUT | unix.EPOLLIN | unix.EPOLLET
	event.Fd = int32(fd)
	if err := unix.EpollCtl(pollerFd, unix.EPOLL_CTL_ADD, fd, &event); err != nil {
		return os.NewSyscallError(fmt.Sprintf("epoll_ctl(%d, ADD, %d, ...)", pollerFd, fd), err)
	}
	return nil
}

func pollEvents(pollerFd int, timeout time.Duration) ([]event, error) {
	var timeoutMS = int(timeout.Nanoseconds() / 1000000)
	var epollEvents [maxEpollEvents]unix.EpollEvent
	nEvents, err := unix.EpollWait(pollerFd, epollEvents[:], timeoutMS)
	if err != nil {
		if err == unix.EINTR {
			return nil, nil
		}
		return nil, os.NewSyscallError("epoll_wait", err)
	}

	var events = make([]event, 0, nEvents)

	for i := 0; i < nEvents; i++ {
		var fd = int(epollEvents[i].Fd)
		var evt = event{Fd: fd, Err: nil}

		errCode, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil {
			evt.Err = os.NewSyscallError("getsockopt", err)
		}
		if errCode != 0 {
			evt.Err = &connectError{unix.Errno(errCode)}
		}
		events = append(events, evt)
	}
	return events, nil
}

func parseSockAddr(addr string) (sAddr unix.Sockaddr, family int, err error) {
	tAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	if ip := tAddr.IP.To4(); ip != nil {
		var addr4 [net.IPv4len]byte
		copy(addr4[:], ip)
		sAddr = &unix.SockaddrInet4{Port: tAddr.Port, Addr: addr4}
		family = unix.AF_INET
		return
	}

	if ip := tAddr.IP.To16(); ip != nil {
		var addr16 [net.IPv6len]byte
		copy(addr16[:], ip)
		sAddr = &unix.SockaddrInet6{Port: tAddr.Port, Addr: addr16}
		family = unix.AF_INET6
		return
	}

	err = &net.AddrError{
		Err:  "unsupported address family",
		Addr: tAddr.IP.String(),
	}
	return
}

func connect(fd int, addr unix.Sockaddr) (success bool, err error) {
	switch serr := unix.Connect(fd, addr); serr {
	case unix.EALREADY, unix.EINPROGRESS, unix.EINTR:
		success = false
		err = nil
	case nil, unix.EISCONN:
		success = true
		err = nil
	case unix.EINVAL:
		if runtime.GOOS == "solaris" {
			success = true
			err = nil
		} else {
			success = false
			err = serr
		}
	default:
		success = false
		err = serr
	}
	return success, err
}
