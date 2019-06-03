package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type imports map[string]string

func (i imports) String() string {
	return "" // irrelevant
}

func (imp *imports) Set(s string) error {
	i := strings.IndexByte(s, '=')
	var name, path string
	if i == -1 {
		path = s
		i = strings.LastIndexByte(s, '/')
		if i > 0 {
			name = s[i+1:]
		} else {
			name = path
		}
	} else {
		name = s[:i]
		path = s[i+1:]
	}

	// FIXME check that name and path have len > 0

	if *imp == nil {
		*imp = imports{name: path}
	} else {
		(*imp)[name] = path
	}
	return nil
}

func main() {
	err := _main()
	if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() > 0 {
		os.Exit(exit.ExitCode())
	} else if err != nil {
		log.Fatal(err)
	}
}

func _main() error {
	var imports imports

	flag.Var(&imports, "i", "import package: [name=]path")
	// TODO declare usage
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("FIXME show usage")
	}
	code := flag.Arg(0)
	args := flag.Args()[1:]

	var src bytes.Buffer
	src.WriteString("package main\n")
	for name, path := range imports {
		fmt.Fprintf(&src, "import %s %q\n", name, path)
	}
	src.WriteString("func main() {\n//line :1\n")
	src.WriteString(code)
	src.WriteString("\n}\n")

	// fmt.Print(src.String())

	f, err := ioutil.TempFile("", "*.go")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	_, err = f.Write(src.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	cmd := exec.Command("go", append([]string{
		"run",
		f.Name(),
	}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// exec.ExitError is handled in caller
	return cmd.Run()
}
