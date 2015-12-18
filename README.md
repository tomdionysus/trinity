![Trinity DB Logo](gfx/trinity_m.png) 

# Trinity DB

Trinity is a concept project for a relational database, designed from the ground up as a cloud system.

## Status

Trinity is pre-alpha - 'concept project'. Don't even think about using this even in development - yet. For how we're going, please see [Progress](docs/progress.md).

## Quick Start

```bash
git clone git@github.com:tomdionysus/trinity.git
cd trinity
go get
bin/make

build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info
```

You can boot other nodes on different ports and have them connect to the cluster like so:

```bash
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -port 13532 -node localhost:13531
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -port 13532 -node localhost:13531
```

## Usage

```bash
trinity-server --ca=<CA_PEM> --cert=<CERT_PEM> [other flags]
```

| Flag                  | Default   | Description                                            |
|:----------------------|-----------|:-------------------------------------------------------|
| -help                 |           | Display command line flags help                        |
| -ca             		|           | Specify the Certificate Authority PEM file             |
| -cert         		|           | Specify the Certificate PEM file                       |
| -loglevel  			| error     | Set the logging level [debug,info,warn,error]          |
| -memcache             | false     | Enable the Memcache interface                          |
| -memcacheport         | 11211     | Set the port for memcache, default 11211               |
| -node                 |           | Specify another Trinity node, e.g. ip_address:port     |

## Documentation

| Document                             | Description                                            |
|:-------------------------------------|:-------------------------------------------------------|
| [Design Goals](docs/design-goals.md) | Design Goals of the project, roadmaps, reference       |
| [Progress](docs/progress.md)         | Project Progress                                       |
| [Encryption](docs/encryption.md)     | Encryption and security guide                          |

## References

* [Golang](https://golang.org)
* [Consistent Hashing](https://en.wikipedia.org/wiki/Consistent_hashing)
* [Memcache Protocol](https://github.com/memcached/memcached/blob/master/doc/protocol.txt)