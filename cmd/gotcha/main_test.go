package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var TestRoot = "t"

func TestRun(t *testing.T) {
	testRoot := filepath.Join(TestRoot, "run")
	if err := os.MkdirAll(testRoot, 0777); err != nil {
		t.Fatal(err)
	}
	newbufs := func() (buf *bytes.Buffer, errbuf *bytes.Buffer) {
		return bytes.NewBufferString(""), bytes.NewBufferString("")
	}
	newopt := func() *option {
		return &option{}
	}

	// TODO: consider to split to subtests and add many cases
	t.Run("valid exit", func(t *testing.T) {
		opt := newopt()
		root := filepath.Join(testRoot, "valid_exit")
		if err := os.MkdirAll(root, 0777); err != nil {
			t.Fatal(err)
		}
		opt.root = root
		if err := ioutil.WriteFile(filepath.Join(root, "hello.txt"), []byte("TODO: hello"), 0666); err != nil {
			t.Fatal(err)
		}
		buf, errbuf := newbufs()
		exp := filepath.Join(root, "hello.txt") + "\n" + "L1:TODO: hello" + "\n\n"

		testf := func() {
			if exit := run(buf, errbuf, opt); exit != ValidExit {
				t.Fatalf("exit=%d errbuf=%s", exit, errbuf)
			}
			if exp != buf.String() {
				t.Errorf("exp=%s\nout=%s", exp, buf)
			}
		}

		t.Log("valid exit")
		testf()

		t.Log("use cache")
		opt.cache = true
		buf.Reset()
		errbuf.Reset()
		testf()
		opt.cache = false

		t.Log("with total")
		opt.total = true
		buf.Reset()
		errbuf.Reset()
		tmp := exp
		exp = exp + "files 1" + "\n" + "lines 1" + "\n" + "errors 0" + "\n"
		testf()
		exp = tmp
		opt.total = false

		t.Log("with sync")
		opt.sync = true
		buf.Reset()
		errbuf.Reset()
		testf()
		opt.sync = false
	})

	t.Run("version", func(t *testing.T) {
		opt := newopt()
		buf, errbuf := newbufs()
		opt.version = true
		exit := run(buf, errbuf, opt)
		if exit != ValidExit {
			t.Error(errbuf)
			t.Error(buf)
		}
		exp := fmt.Sprintln(Name + " version " + Version)
		if buf.String() != exp {
			t.Errorf("exp=%s but out=%s err=%s", exp, buf, errbuf)
		}
	})

	t.Run("out", func(t *testing.T) {
		opt := newopt()
		buf, errbuf := newbufs()

		root := filepath.Join(testRoot, "out")
		if err := os.MkdirAll(root, 0777); err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)
		opt.root = root

		path := filepath.Join(root, "text.txt")
		contents := "TODO: hello"
		exp := path + "\n" + "L1:TODO: hello" + "\n\n"
		if err := ioutil.WriteFile(path, []byte(contents), 0666); err != nil {
			t.Fatal(err)
		}
		outpath := filepath.Join(root, "tmp.log")
		opt.out = outpath

		// out to tmp.log
		if exit := run(buf, errbuf, opt); exit != ValidExit {
			t.Fatalf("exit=%d errbuf=%s", exit, errbuf)
		}
		b, err := ioutil.ReadFile(outpath)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != exp {
			t.Errorf("exp=%s out=%s", exp, buf)
		}

		// reject override
		buf.Reset()
		errbuf.Reset()
		if exit := run(buf, errbuf, opt); exit != ErrInitialize {
			t.Fatalf("[reject override] expected exit=%d exit=%d errbuf=%s opt=%#v", ErrInitialize, exit, errbuf, opt)
		}

		// use force
		opt.force = true
		buf.Reset()
		errbuf.Reset()
		if exit := run(buf, errbuf, opt); exit != ValidExit {
			t.Fatalf("[use force] exit=%d errbuf=%s opt=%#v", exit, errbuf, opt)
		}

		// reject directory
		buf.Reset()
		errbuf.Reset()
		dir := filepath.Join(root, "dir")
		if err := os.Mkdir(dir, 0777); err != nil {
			t.Fatal(err)
		}
		opt.out = dir
		opt.force = true
		if exit := run(buf, errbuf, opt); exit != ErrInitialize {
			t.Fatalf("[reject directory] expected exit=%d exit=%d errbuf=%s opt=%#v", ErrInitialize, exit, errbuf, opt)
		}
	})

	t.Run("specify file", func(t *testing.T) {
		root := filepath.Join(testRoot, "specify_file")
		if err := os.MkdirAll(root, 0777); err != nil {
			t.Fatal(err)
		}
		opt := newopt()
		buf, errbuf := newbufs()
		path := filepath.Join(root, "specify.txt")
		if err := ioutil.WriteFile(path, []byte("TODO: hello"), 0666); err != nil {
			t.Fatal(err)
		}
		exp := path + "\n" + "L1:TODO: hello" + "\n\n"
		opt.root = path
		if exit := run(buf, errbuf, opt); exit != ValidExit {
			t.Errorf("exit=%d errbuf=%s opt=%#v", exit, errbuf, opt)
		}
		if exp != buf.String() {
			t.Errorf("exp=%#v out=%#v opt=%#v", exp, buf, opt)
		}
	})
}
