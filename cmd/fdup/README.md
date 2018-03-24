fdup
====
fdup is tool of find duplicate contents  
actually is compare hash

Usage:
------
display files with have same contents

compare two files
```sh
fdup /path/file1 /path/file2
```

check recursive
```sh
fdup /path/dir
```

specify file and directory
```sh
fdup /path/file /path/dir
```

with verbose
```sh
fdup -verbose /path/file /path/dir
```

Install:
--------
```sh
go get -v -u github.com/yaeshimo/go-utils/cmd/fdup
```

TODO:
-----
multithread

License:
--------
MIT
