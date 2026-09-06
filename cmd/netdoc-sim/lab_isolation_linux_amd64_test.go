package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// A fresh process makes this irreversible filter local to the test child. It
// kills attempted sockets (including netlink/interface queries), file opens
// (including resolv.conf, hosts and routing files), and subprocess execution.
// No network syscall reaches the kernel's network implementation.
func TestLabKernelIsolation(t *testing.T) {
	if mode := os.Getenv("NETDOC_LAB_ISOLATION_HELPER"); mode != "" {
		runtime.LockOSThread()
		filter := []unix.SockFilter{{Code: unix.BPF_LD | unix.BPF_W | unix.BPF_ABS, K: 0}}
		for _, nr := range []uint32{unix.SYS_SOCKET, unix.SYS_SOCKETPAIR, unix.SYS_CONNECT, unix.SYS_OPEN, unix.SYS_CREAT, unix.SYS_OPENAT, unix.SYS_OPENAT2, unix.SYS_EXECVE, unix.SYS_EXECVEAT} {
			filter = append(filter, unix.SockFilter{Code: unix.BPF_JMP | unix.BPF_JEQ | unix.BPF_K, K: nr, Jf: 1}, unix.SockFilter{Code: unix.BPF_RET | unix.BPF_K, K: unix.SECCOMP_RET_KILL_PROCESS})
		}
		filter = append(filter, unix.SockFilter{Code: unix.BPF_RET | unix.BPF_K, K: unix.SECCOMP_RET_ALLOW})
		program := unix.SockFprog{Len: uint16(len(filter)), Filter: &filter[0]}
		if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
			t.Fatal(err)
		}
		_, _, errno := unix.RawSyscall(unix.SYS_SECCOMP, unix.SECCOMP_SET_MODE_FILTER, unix.SECCOMP_FILTER_FLAG_TSYNC, uintptr(unsafe.Pointer(&program)))
		if errno != 0 {
			t.Fatal(errno)
		}
		switch mode {
		case "socket":
			_, _ = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
		case "file":
			_, _ = os.Open("/etc/resolv.conf")
		case "lab":
			os.Exit(run([]string{"lab", "run", "--all", "--json"}, os.Stdout, os.Stderr))
		}
		os.Exit(99) // forbidden operation unexpectedly survived
	}
	for _, mode := range []string{"lab", "socket", "file"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLabKernelIsolation$")
			// Set these before startup: Go CPU polling and glibc's lazy malloc
			// arena sizing can otherwise open CPU topology files after filtering.
			cmd.Env = append(os.Environ(), "GOMAXPROCS=2", "MALLOC_ARENA_MAX=2", "NETDOC_LAB_ISOLATION_HELPER="+mode, "HTTPS_PROXY=http://unresolvable.invalid:1", "ALL_PROXY=socks5://unresolvable.invalid:1")
			var out, errOut bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &errOut
			err := cmd.Run()
			if mode == "lab" {
				if cmd.ProcessState.ExitCode() != exitMismatch || !bytes.Contains(out.Bytes(), []byte(`"Scenario": "tls-http-no-response"`)) {
					t.Fatalf("isolated lab: %v %s", err, errOut.String())
				}
				var ordinary bytes.Buffer
				if code := run([]string{"lab", "run", "--all", "--json"}, &ordinary, &errOut); code != exitMismatch || !bytes.Equal(out.Bytes(), ordinary.Bytes()) {
					t.Fatal("kernel isolation changed serialized evidence")
				}
			} else if cmd.ProcessState.Sys().(syscall.WaitStatus).Signal() != syscall.SIGSYS {
				t.Fatalf("negative control was not killed: %v %s", err, errOut.String())
			}
		})
	}
}
