![Trinity DB Logo](gfx/trinity_m.png) 

# Trinity DB

Trinity is a concept project for a relational database, designed from the ground up as a cloud system.

## Design goals

* Inherently multi master - connect to any node
* Automatic replication and sharding
* Optional Direct block level access, no filesystem
* Multiple-mode consistency (fire-and-forget, eventual, full)
* Soft clusters, add/remove nodes at any time
* Zero configuration
* Capacity scales by adding nodes
* ANSI SQL compatible

## Language

Trinity is written in [Golang](https://golang.org).

## Constraints on SQL

* Tables must have an 8-byte random primary key.

## Progress

## TODO

* 