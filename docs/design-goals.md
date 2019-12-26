![Trinity DB Logo](../gfx/trinity_m.png) 

# Trinity DB - Design Goals

* Distributed Architecture - no masters/replicas/slaves, read/write to any node
* ANSI SQL92 compatible
* Built-in fast, distributed Key-Value Store
* Automatic replication and sharding
* Distributed Queries 
* Multi-mode consistency - per write choose: fire-and-forget, eventual, full
* Soft clusters - add/remove nodes at any time
* Capacity scales by adding nodes
* Optional Direct block level access, no filesystem
* All connections encrypted with TLS
* Zero configuration