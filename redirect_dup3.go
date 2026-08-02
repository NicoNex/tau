//go:build linux && (arm64 || riscv64 || loong64 || mips64 || mips64le)

package tau

import "syscall"

// The newer Linux architectures never got the dup2 system call - arm64 among
// them - and offer only dup3, which is dup2 with a flags argument. With no
// flags and two different descriptors the two are the same call.
func dupTo(oldfd, newfd int) error {
	return syscall.Dup3(oldfd, newfd, 0)
}
