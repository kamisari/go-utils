package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func ExampleSync() {
	/// init
	// test directory
	testRoot := "t"
	testRoot = filepath.Join(testRoot, "example")
	if err := os.MkdirAll(testRoot, 0777); err != nil {
		panic(err)
	}
	// test files
	var (
		sameFilesContents = "hello world"
		sameFiles         = []string{
			filepath.Join(testRoot, "same_one.txt"),
			filepath.Join(testRoot, "same_two.txt"),
		}

		uniqueFilesContents = "unique"
		uniqueFile          = filepath.Join(testRoot, "unique.txt")
	)

	write := func(path string, b []byte) {
		if err := ioutil.WriteFile(path, b, 0666); err != nil {
			panic(err)
		}
	}
	write(sameFiles[0], []byte(sameFilesContents))
	write(sameFiles[1], []byte(sameFilesContents))
	write(uniqueFile, []byte(uniqueFilesContents))

	files := []string{uniqueFile, sameFiles[0], sameFiles[1]}
	Sync(os.Stdout, os.Stderr, DefaultHashAlgorithm, files)
	// Output:
	// Used hash algorithm: "sha512_256"
	// Conflicted hash [0ac561fac838104e3f2e4ad107b4bee3e938bf15f2b15f009ccccd61a913f017]
	// 	"t/example/same_one.txt"
	// 	"t/example/same_two.txt"
}

func ExampleAsync() {
	/// init
	// test directory
	testRoot := "t"
	testRoot = filepath.Join(testRoot, "example_async")
	if err := os.MkdirAll(testRoot, 0777); err != nil {
		panic(err)
	}
	// test files
	var (
		sameFilesContents = "hello world"
		sameFiles         = []string{
			filepath.Join(testRoot, "same_one.txt"),
			filepath.Join(testRoot, "same_two.txt"),
		}

		uniqueFilesContents = "unique"
		uniqueFile          = filepath.Join(testRoot, "unique.txt")
	)

	write := func(path string, b []byte) {
		if err := ioutil.WriteFile(path, b, 0666); err != nil {
			panic(err)
		}
	}
	write(sameFiles[0], []byte(sameFilesContents))
	write(sameFiles[1], []byte(sameFilesContents))
	write(uniqueFile, []byte(uniqueFilesContents))

	files := []string{uniqueFile, sameFiles[0], sameFiles[1]}
	Async(os.Stdout, os.Stderr, DefaultHashAlgorithm, files)
	// Unordered output:
	// Used hash algorithm: "sha512_256"
	// Conflicted hash [0ac561fac838104e3f2e4ad107b4bee3e938bf15f2b15f009ccccd61a913f017]
	// 	"t/example_async/same_one.txt"
	// 	"t/example_async/same_two.txt"
}
