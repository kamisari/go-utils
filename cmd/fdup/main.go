package main

import (
	"crypto/md5"
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// base information
const (
	Name                 = "fdup"
	Version              = "0.1.0"
	DefaultHashAlgorithm = "sha512_256"
)

type option struct {
	version  bool
	verbose  bool
	hash     string
	fullpath bool

	// TODO: impl flags
	// 1. list available hash algorithm
	// 2. use multithread

	// TODO: consider
	async bool
}

var opt = &option{}
var errLogger = log.New(ioutil.Discard, "["+Name+"]:", log.Lshortfile)

func init() {
	flag.BoolVar(&opt.version, "version", false, "show version")
	flag.BoolVar(&opt.verbose, "verbose", false, "verbose")
	flag.StringVar(&opt.hash, "hash", DefaultHashAlgorithm, "specify use hash algorithm")
	flag.BoolVar(&opt.fullpath, "fullpath", false, "output with fullpath")

	// TODO: consider
	flag.BoolVar(&opt.async, "async", false, "async calculate")

	log.SetPrefix("[" + Name + "]:")
	log.SetOutput(ioutil.Discard)
}

// Sync run sync
func Sync(stdout, stderr io.Writer, usehash string, targets []string) int {
	var checker hash.Hash
	switch usehash {
	case "md5":
		checker = md5.New()
	case "sha512_256":
		checker = sha512.New512_256()
	default:
		fmt.Fprintln(stderr, "invalid hash algorithm:", usehash)
		return 1
	}

	type dup struct {
		isDupl bool
		sum    []byte
		path   []string
	}

	// key=fmt.Sprintf("%x", checker.Sum(nil))
	hashMap := make(map[string]*dup)
	// key=FilePath for avoid duplicate check
	avoidMap := make(map[string]bool)
	for _, root := range targets {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() {
				if avoidMap[path] {
					return nil
				}
				avoidMap[path] = true

				// hash check
				f, err := os.Open(path)
				if err != nil {
					errLogger.Println(err)
					return nil
				}
				defer f.Close()
				checker.Reset()
				if _, err := io.Copy(checker, f); err != nil {
					errLogger.Println(err)
					return nil
				}
				key := fmt.Sprintf("%x", checker.Sum(nil))
				if d, ok := hashMap[key]; ok {
					d.isDupl = true
					d.path = append(d.path, path)
				} else {
					hashMap[key] = &dup{isDupl: false, sum: checker.Sum(nil), path: []string{path}}
				}

				// for verbose
				log.Printf("checked: %q [%s]", path, key)
			}
			return nil
		})
		if err != nil {
			errLogger.Println(err)
		}
	}

	fmt.Fprintf(stdout, "Used hash algorithm: %q\n", usehash)
	for _, d := range hashMap {
		if d.isDupl {
			fmt.Fprintf(stdout, "Conflicted hash [%x]\n", d.sum)
			for _, s := range d.path {
				fmt.Fprintf(stdout, "\t%q\n", s)
			}
		}
	}
	return 0
}

// Async run async
// TODO: consider
func Async(stdout, stderr io.Writer, usehash string, targets []string) (exit int) {
	var newChecker func() hash.Hash
	switch usehash {
	case "md5":
		newChecker = func() hash.Hash { return md5.New() }
	case "sha512_256":
		newChecker = func() hash.Hash { return sha512.New512_256() }
	default:
		fmt.Fprintln(stderr, "not supported hash algorithm:", usehash)
		return 1
	}

	type dup struct {
		isDupl bool
		sum    []byte
		path   []string
	}
	// key=fmt.Sprintf("%x", checker.Sum(nil))
	hashMap := make(map[string]*dup)

	/// make hashMap
	wg := new(sync.WaitGroup)
	queue := make(chan string, 128)
	type result struct {
		path string
		hash []byte
	}
	resch := make(chan *result, 32)
	func() {
		/// push results
		go func() {
			for {
				if res := <-resch; res == nil {
					exit = 1
				} else {
					key := fmt.Sprintf("%x", res.hash)
					if d, ok := hashMap[key]; ok {
						d.isDupl = true
						d.path = append(d.path, res.path)
					} else {
						hashMap[key] = &dup{isDupl: false, sum: res.hash, path: []string{res.path}}
					}
					// for verbose
					log.Printf("checked: %q [%s]", res.path, key)
				}
				wg.Done()
			}
		}()

		/// go worker
		n := runtime.NumCPU()
		if n < 1 {
			n = 1
		}
		for i := 0; i < n; i++ {
			go func() {
				checker := newChecker()
				for {
					path := <-queue
					func() {
						f, err := os.Open(path)
						if err != nil {
							errLogger.Println(err)
							resch <- nil
							return
						}
						defer f.Close()
						checker.Reset()
						if _, err := io.Copy(checker, f); err != nil {
							errLogger.Println(err)
							resch <- nil
							return
						}
						resch <- &result{path: path, hash: checker.Sum(nil)}
					}()
				}
			}()
		}
	}()

	/// walk directory
	// key=FilePath for avoid duplicate check
	avoidMap := make(map[string]bool)
	for _, root := range targets {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() {
				if avoidMap[path] {
					return nil
				}
				avoidMap[path] = true
				wg.Add(1)
				queue <- path
			}
			return nil
		})
		if err != nil {
			errLogger.Println(err)
			exit = 1
		}
	}

	wg.Wait()

	/// output
	fmt.Fprintf(stdout, "Used hash algorithm: %q\n", usehash)
	for _, d := range hashMap {
		if d.isDupl {
			fmt.Fprintf(stdout, "Conflicted hash [%x]\n", d.sum)
			for _, s := range d.path {
				fmt.Fprintf(stdout, "\t%q\n", s)
			}
		}
	}
	return exit
}

func main() {
	flag.Parse()
	if opt.version {
		fmt.Fprintf(os.Stdout, "%s version %s\n", Name, Version)
		os.Exit(0)
	}
	targets := flag.Args()
	if opt.fullpath {
		var newtargets []string
		for _, path := range targets {
			if abs, err := filepath.Abs(path); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			} else {
				newtargets = append(newtargets, abs)
			}
		}
		targets = newtargets
	}
	if opt.verbose {
		log.SetOutput(os.Stdout)
	}
	errLogger.SetOutput(os.Stderr)

	// TODO: consider
	if opt.async {
		os.Exit(Async(os.Stdout, os.Stderr, opt.hash, targets))
	} else {
		os.Exit(Sync(os.Stdout, os.Stderr, opt.hash, targets))
	}
}
