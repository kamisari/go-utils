package main

import (
	"path/filepath"
	"testing"
)

func Test_makeReadme(t *testing.T) {
	testDir := filepath.Join("t", "TitleForReadme")
	author := "makereadme"
	exp := `TitleForReadme
==============

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
makereadme
`

	if out := makeReadme(testDir, author); out != exp {
		t.Fatalf("exp %s but out %s", exp, out)
	}
}
