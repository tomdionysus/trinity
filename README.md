![Trinity DB Logo](gfx/trinity_m.png) 

# Trinity DB

Trinity is a concept project for a relational database, designed from the ground up as a cloud system.

## Status

Trinity is pre-alpha - 'concept project'. Don't even think about using this even in development - yet.

## Design goals

* Distributed Architecture - no masters/replicas/slaves, read/write to any node
* ANSI SQL92 compatible
* Built-in fast, distrubuted Key-Value Store
* Automatic replication and sharding
* Distributed Queries 
* Multi-mode consistency - per write, choose fire-and-forget, eventual, full
* Soft clusters - add/remove nodes at any time
* Capacity scales by adding nodes
* Optional Direct block level access, no filesystem
* All connections encrypted with TLS
* Zero configuration

## Language

Trinity is written in [Golang](https://golang.org).

## Progress

* Command Line flags
* TLS Layer Prototype
* GOB streaming between servers
* Heartbeats

## TODO

* Peer Swapping
* Autoconnect to all available nodes
* Proxying GOBs, Data
* Integrate consistenthash
* Disk-based key/value store 
* Expose Key/value store