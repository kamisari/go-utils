package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	name    = "gits"
	version = "0.5.1"
)

// Default values
// TODO: separate to conf_*.go
var (
	CandidateConfPaths = func() (s []string) {
		u, err := user.Current()
		if err != nil {
			return
		}
		if u.HomeDir != "" {
			s = append(s, filepath.Join(u.HomeDir, ".gits.json"))
		}
		return
	}()
	EditorWithArgs = []string{"vim", "--"}
)

type option struct {
	version bool
	conf    string
	exec    string

	match string

	edit  bool
	add   string
	rm    string
	prune bool
	dir   string

	list           bool
	listRepo       bool
	listRepoFull   bool
	listAlias      bool
	listCandidates bool

	template bool
}

var opt = &option{}

func init() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("[" + name + "]:")
	flag.BoolVar(&opt.version, "version", false, "show version")
	flag.StringVar(&opt.conf, "config", "", "specify path to configuration JSON format files")
	flag.StringVar(&opt.exec, "exec", "git", "specify executable command name")

	flag.StringVar(&opt.match, "match", "", "for pick any repostories with regexp RE2")

	flag.BoolVar(&opt.edit, "edit", false, "edit config")
	flag.StringVar(&opt.add, "add", "", "specify path to directory for add to configuration files")
	flag.StringVar(&opt.rm, "rm", "", "specify key to remove from configuration file")
	flag.BoolVar(&opt.prune, "prune", false, "prune invalid worktree from configuration file")
	flag.StringVar(&opt.dir, "dir", "", "specify repository then to show path to worktree")

	flag.BoolVar(&opt.list, "list", false, "show content of configuration file")
	flag.BoolVar(&opt.listRepo, "list-repo", false, "list repositories")
	flag.BoolVar(&opt.listRepoFull, "list-repo-full", false, "list repositories with full path")
	flag.BoolVar(&opt.listAlias, "list-alias", false, "list alias")
	flag.BoolVar(&opt.listCandidates, "list-candidates", false, "list candidate paths to the configuration file")

	flag.BoolVar(&opt.template, "template", false, "show configuration template")
}

// Edit edit configuration file
func Edit(w, errw io.Writer, r io.Reader, path string) error {
	if len(EditorWithArgs) < 1 {
		return fmt.Errorf("invalid [EditorWithArgs]: %v", EditorWithArgs)
	}
	editor := exec.Command(EditorWithArgs[0], append(EditorWithArgs[1:], path)...)
	editor.Stdout = w
	editor.Stderr = errw
	editor.Stdin = r
	if _, err := fmt.Fprintln(w, editor.Args); err != nil {
		return err
	}
	return editor.Run()
}

func main() {
	// TODO: consider to split to functions from flags
	// 1. run
	// 2. check error
	// 3. output err or valid message

	// TODO: consider
	validateArgs := func(n int) {
		if flag.NArg() != n {
			flag.PrintDefaults()
			log.Fatalf("invalid arguments %v\n", flag.Args())
		}
	}

	flag.Parse()

	if opt.version {
		validateArgs(0)
		fmt.Fprintf(os.Stdout, "%s version %s\n", name, version)
		return
	}
	if opt.conf == "" {
		for _, path := range CandidateConfPaths {
			if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
				opt.conf = path
			}
		}
	}
	if opt.edit {
		validateArgs(0)
		if err := Edit(os.Stdout, os.Stderr, os.Stdin, opt.conf); err != nil {
			log.Fatal(err)
		}
		return
	}
	if opt.template {
		validateArgs(0)
		b, err := Template()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", string(b))
		return
	}
	if opt.listCandidates {
		validateArgs(0)
		fmt.Fprintf(os.Stdout, "Candidates:\n")
		for i, s := range CandidateConfPaths {
			fmt.Fprintf(os.Stdout, "\t%d. %s\n", i+1, s)
		}
		return
	}

	gits, err := ReadJSON(opt.conf)
	if err != nil {
		log.Fatal(err)
	}
	if opt.match != "" {
		if err := gits.RemoveMatchRepositories(opt.match); err != nil {
			log.Fatal(err)
		}
	}
	switch {
	case opt.add != "":
		validateArgs(0)
		root, err := GetGitToplevel(opt.add)
		if err != nil {
			log.Fatal(err)
		}
		if err := gits.AddRepository("", root); err != nil {
			log.Fatal(err)
		}
		if err := gits.Update(); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Appended Repositories:\n\t[%s]\nCurrent List:\n", root)
		if err := gits.FprintIndent(os.Stdout, "", "\t"); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Updated:\n\t[%s]\n", gits.Path())
	case opt.rm != "":
		validateArgs(0)
		if err := gits.RemoveRepository(opt.rm); err != nil {
			log.Fatal(err)
		}
		if err := gits.Update(); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Removed Repositories:\n\t[%s]\nCurrent List:\n", opt.rm)
		if err := gits.FprintIndent(os.Stdout, "", "\t"); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Updated:\n\t[%s]\n", gits.Path())
	case opt.prune:
		validateArgs(0)
		if removed, err := gits.Prune(); err != nil {
			log.Fatal(err)
		} else if len(removed) != 0 {
			fmt.Fprintf(os.Stdout, "Pruned:\n\t\"%s\"\n", strings.Join(removed, "\n\t"))
		} else {
			fmt.Fprintf(os.Stdout, "Already clean\n")
			return
		}
		if err := gits.Update(); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Current List:\n")
		if err := gits.FprintIndent(os.Stdout, "", "\t"); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Updated:\n\t[%s]\n", gits.Path())
	case opt.dir != "":
		validateArgs(0)
		if repo, ok := gits.Repositories[opt.dir]; ok {
			fmt.Fprintf(os.Stdout, "%s\n", repo.WorkTree)
		} else {
			log.Fatalf("not exists %s in %s\n", opt.dir, gits.Path())
		}
	case opt.list:
		validateArgs(0)
		if err := gits.FprintIndent(os.Stdout, "", "\t"); err != nil {
			log.Fatal(err)
		}
	case opt.listRepo:
		validateArgs(0)
		gits.ListRepositories(os.Stdout)
	case opt.listRepoFull:
		validateArgs(0)
		gits.ListRepositoriesFull(os.Stdout)
	case opt.listAlias:
		validateArgs(0)
		if err := gits.ListAlias(os.Stdout, opt.exec); err != nil {
			log.Fatal(err)
		}
	default:
		var alias string
		if n := flag.NArg(); n == 1 {
			alias = flag.Arg(0)
		} else {
			fmt.Fprintf(os.Stderr, "invalid arguments: %v\n", flag.Args())
			if err := gits.ListAlias(os.Stdout, opt.exec); err != nil {
				log.Fatal(err)
			}
			os.Exit(1)
		}
		if err := gits.Run(os.Stdout, os.Stderr, os.Stdin, opt.exec, alias); err != nil {
			log.Fatal(err)
		}
	}
}
