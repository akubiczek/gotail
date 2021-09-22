package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var numLines int

func initFlags() {
	flag.IntVar(&numLines, "n", 10, "First line to display, use negative number for line relative to the end of file")
	flag.Parse()
}

func validateArgs(args []string) error {
	if len(args) < 2 {
		println("usage:", filepath.Base(args[0]), "[-n #] file")
		return errors.New("bad command line arguments")
	}

	return nil
}

func seekToPositiveLine(file *os.File, lineNumber int) error {
	var _, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	const BufferSize int = 8 * 1024
	buf := make([]byte, BufferSize)
	currentLine := 0
	lineSep := []byte{'\n'}
	for {
		c, err := file.Read(buf)
		currentLine += bytes.Count(buf[:c], lineSep)
		linesDiff := currentLine - lineNumber + 1

		if linesDiff > 0 {
			for i := c; i > 0; i-- {
				if buf[i] == lineSep[0] {
					linesDiff--
					if linesDiff == 0 {
						_, err = file.Seek(int64(i-c+1), io.SeekCurrent)
						return err
					}
				}
			}
		}

		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func seekToNegativeLine(file *os.File, lineNumber int) error {
	const BufferSize int64 = 8 * 1024

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() < BufferSize {
		_, err = file.Seek(0, io.SeekStart)
	} else {
		_, err = file.Seek(-BufferSize, io.SeekEnd)
	}

	if err != nil {
		return err
	}

	buf := make([]byte, BufferSize)
	currentLine := 0
	lineSep := []byte{'\n'}
	var currentOffset int64 = BufferSize
	var step = 1
	var extra = 1

	for {
		c, err := file.Read(buf)

		if currentOffset != BufferSize {
			//beginning of file has been reached, so we cannot count all bytes in buffer but remaining only
			c = int(currentOffset)
			extra = 0
		}

		currentLine -= bytes.Count(buf[:c], lineSep)
		linesDiff := -(currentLine - lineNumber - 1)
		if linesDiff > 0 {
			for i := 0; i < c; i++ {
				if buf[i] == lineSep[0] {
					linesDiff--
					if linesDiff == 0 {
						_, err = file.Seek(-int64(c-i-extra), io.SeekCurrent)
						return err
					}
				}
			}
		}

		if fileInfo.Size() < BufferSize || currentOffset != BufferSize {
			_, err = file.Seek(0, io.SeekStart)
			return nil
		}

		_, err = file.Seek(-int64(c), io.SeekCurrent)
		if err != nil {
			return err
		}

		seekOffset := BufferSize
		currentOffset, err = file.Seek(0, io.SeekCurrent)

		if currentOffset < seekOffset {
			//beginning of file has been reached
			seekOffset = currentOffset
		} else {
			currentOffset = BufferSize
		}
		_, err = file.Seek(-seekOffset, io.SeekCurrent)
		step++

		if err != nil {
			return err
		}
	}
}

func seekToLine(file *os.File, lineNumber int) error {

	if lineNumber > 0 {
		return seekToPositiveLine(file, lineNumber)
	} else if lineNumber < 0 {
		return seekToNegativeLine(file, lineNumber)
	}

	return nil
}

func command() error {

	if fileName := flag.Arg(0); fileName != "" {

		file, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		err = seekToLine(file, numLines)
		if err != nil {
			return err
		}

		io.Copy(os.Stdout, file)
	}

	return nil
}

func main() {

	initFlags()
	if err := validateArgs(os.Args[0:]); err == nil {
		if err := command(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
