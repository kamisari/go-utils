package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// default contents
const (
	/// main.go
	ContentsForMain = `package main

func main() {
	println("hello world")
}`

	/// main_test.go
	ContentsForMainTest = `package main

import (
	"testing"
)

func TestRun(t *testing.T) {
	t.Fatal()
}`

	/// .gitignore
	ContentsForGitignore = `*.log
*.prof`
)

// titleForReadme is retrun string with underline
func makeReadme(dir string, author string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	base := filepath.Base(abs)
	var under string
	for range base {
		under += "="
	}

	/// README.md
	var readme = fmt.Sprintf("%s\n%s", base, under) + `

Usage:
------

Requirements:
-------------

Install:
--------

License:
--------

Author:
-------
` + fmt.Sprintln(author)

	return readme
}

func bin(args []string) error {
	flagSet := flag.NewFlagSet("bin", flag.ExitOnError)
	// TODO: append flags
	flagSet.Parse(args)

	// TODO: consider
	if flagSet.NArg() != 1 {
		return fmt.Errorf("invalid arguments: %v", flagSet.Args())
	}
	initdir := flagSet.Arg(0)

	if err := os.Mkdir(initdir, 0777); err != nil {
		log.Fatal(err)
	}
	if err := os.Chdir(initdir); err != nil {
		log.Fatal(err)
	}

	/// main.go
	if err := ioutil.WriteFile("main.go", []byte(ContentsForMain), 0666); err != nil {
		return err
	}

	/// main_test.go
	if err := ioutil.WriteFile("main_test.go", []byte(ContentsForMainTest), 0666); err != nil {
		return err
	}

	/// git
	var author string
	if err := func() error {
		if _, err := exec.LookPath("git"); err != nil {
			return nil
		}
		if err := exec.Command("git", "init").Run(); err != nil {
			return err
		}
		/// .gitignore
		if err := ioutil.WriteFile(".gitignore", []byte(ContentsForGitignore), 0666); err != nil {
			return err
		}
		/// author
		if b, err := exec.Command("git", "config", "user.name").Output(); err != nil {
			return err
		} else {
			author = strings.TrimSpace(string(b))
		}
		return nil
	}(); err != nil {
		return err
	}

	/// README.md
	// TODO: author, consider to use: `git config user.name`
	if err := ioutil.WriteFile("README.md", []byte(makeReadme(initdir, author)), 0666); err != nil {
		return err
	}

	return nil
}

// SubCommands map for declaration of subcommands
var SubCommands = make(map[string]func([]string) error)

func init() {
	SubCommands["bin"] = bin
	//SubCommands["lib"] = lib
	SubCommands["list"] = func(args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("invalid arguments: %v", args)
		}
		var str string
		for key := range SubCommands {
			str += fmt.Sprintln(key)
		}
		_, err := fmt.Print(str)
		return err
	}
}

func main() {
	list := flag.Bool("list", false, "list subcommands")
	flag.Parse()
	if n := flag.NArg(); n == 0 && !*list {
		flag.Usage()
		log.Fatal("expected arguments")
	}

	if f, ok := SubCommands[flag.Arg(0)]; ok {
		if err := f(flag.Args()[1:]); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatalf("invalid arguments: %v", flag.Args())
	}
}
