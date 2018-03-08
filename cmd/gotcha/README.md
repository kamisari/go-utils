# gotcha
--------
cli tool for gathering of "TODO: " from files


## Installation:
----------------
```
go get -v -u github.com/kamisari/go-utils/cmd/gotcha
```

## Example:
-----------
- `gotcha` recursive check from current directory
- `gotcha /path/dir` or `gotcha -root /path/dir` specify root
- `gotcha -word "func "` specify target word, default is "TODO: "
- `gotcha -out /path/log` specify output

- `gotcha -help` print help

## Licence:
-----------
MIT
