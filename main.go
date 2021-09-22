package main

import (
	"bufio"
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
	flag.IntVar(&numLines, "n", 10, "Number of lines to display")
	flag.Parse()
}

func validateArgs(args []string) error {
	if len(args) < 2 {
		println("usage:", filepath.Base(args[0]), "[-n #] file")
		return errors.New("bad command line arguments")
	}

	return nil
}

func countLines(fileName string) (int, error) {
	var r, err = os.Open(fileName)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count + 1, nil

		case err != nil:
			return count, err
		}
	}
}

func command() error {
	var in io.Reader
	var linesCount int
	var err error

	if fileName := flag.Arg(0); fileName != "" {

		linesCount, err = countLines(fileName)
		if err != nil {
			return err
		}

		file, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		in = file
	}

	var startLine = linesCount - numLines

	scanner := bufio.NewScanner(in)

	for i := 0; i < linesCount; i++ {
		if !scanner.Scan() {
			break
		}

		if i >= startLine {
			fmt.Println(scanner.Text())
		}
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
