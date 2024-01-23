//go:build windows

package tau

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/windows"
)

func redirectStdout(w io.Writer) {
	pr, pw, err := os.Pipe()
	if err != nil {
		fmt.Println("Error creating pipe:", err)
		return
	}
	// Set the pipe writer as the stdout
	var stdHandle windows.Handle
	err = windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(pw.Fd()), windows.CurrentProcess(), &stdHandle, 0, true, windows.DUPLICATE_SAME_ACCESS)
	if err != nil {
		fmt.Println("Error duplicating handle:", err)
		return
	}
	err = windows.SetStdHandle(windows.STD_OUTPUT_HANDLE, stdHandle)
	if err != nil {
		fmt.Println("Error setting stdout:", err)
		return
	}

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
