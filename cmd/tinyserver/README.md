tinyserver
==========
tiny file server

Usage:
------
Serve
```sh
cd ${serv_root}
tinyserver
# default listen http://127.0.0.1:8080
# push ctrl-c to stop
```

After serve
```sh
# in another terminal
curl http://127.0.0.1:8080
```

Install:
--------
```sh
go get -v -u github.com/kamisari/go-utils/cmd/tinyserver
```

License:
--------
MIT
