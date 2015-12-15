![Trinity DB Logo](gfx/trinity_m.png) 

# Trinity DB

Trinity is a concept project for a relational database, designed from the ground up as a cloud system.

## Status

Trinity is pre-alpha - 'concept project'. Don't even think about using this even in development - yet.

* [Design Goals](docs/design-goals.md)
* [Progress](docs/progress.md)

## Building

Clone the repo, and use the `bin/make` script:

```bash
git clone git@github.com:tomdionysus/trinity.git
cd trinity
bin/make
```

Trinity compiles to a single binary, `build/trinity-server`.

## Usage

```bash
trinity-server --ca=<CA_PEM> --cert=<CERT_PEM> [flags]
```

| Flag                  | Default        | Description                                       |
|:----------------------|----------------|:--------------------------------------------------|
| -help                 | n/a            | Display command line flags help                   |
| -ca             		| ca.pem         | Specify the Certificate Authority PEM file        |
| -cert         		| cert.pem       | Specify the Certificate PEM file                  |
| -loglevel  			| error          | Set the logging level [debug,info,warn,error]     |
| -memcache             | false          | Enable the Memcache interface                     |
| -memcacheport         | 11211          | Set the port for memcache, default 11211          |


## References

* [Golang](https://golang.org)