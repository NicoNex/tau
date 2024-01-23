//go:build !windows

package tau

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

func redirectStdout(w io.Writer) {
	pr, pw, err := os.Pipe()
	if err != nil {
		fmt.Println("Error creating pipe:", err)
		return
	}
	// Set the pipe writer as the stdout
	syscall.Dup2(int(pw.Fd()), syscall.Stdout)

	go func() {
		var buf = make([]byte, 4096)

		for {
			n, err := pr.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Println("error reading from pipe:", err)
				}
				break
			}

			// Write the captured output to the provided writer
			_, err = w.Write(buf[:n])
			if err != nil {
				fmt.Println("error writing to writer:", err)
				break
			}
		}
		pr.Close()
	}()
}
