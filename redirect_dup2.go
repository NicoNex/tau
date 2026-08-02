//go:build !windows && !(linux && (arm64 || riscv64 || loong64 || mips64 || mips64le))

package tau

import "syscall"

// dupTo points one file descriptor at another - here, stdout at a pipe. This
// is the platforms whose syscall package has Dup2: macOS, the BSDs, and Linux
// on the architectures that kept the dup2 system call.
func dupTo(oldfd, newfd int) error {
	return syscall.Dup2(oldfd, newfd)
}
