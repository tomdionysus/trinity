![Trinity DB Logo](../gfx/trinity_m.png) 

# Trinity DB - Progress

## Overview

As of late Dec 2015, Trinity is pre-alpha. The basic node infrastructure is mostly in place, including the server binary, the TLS layer, the ability to listen for cluster connections and connect to other nodes. There is also a single node Memcache server that supports a couple of the basic protocol commands (set with flags/expiry, get, delete). 

The next step will be to have the nodes communicate with each other and swap node lists, with auto-connection to all available nodes. The idea is that every node will be ideally connected to every other node.

After that, each node should support a consistent hashing distribution circle of its own, and swap its distribution to other nodes. Then, the KV store will be adapted to store data on the appropriate node, plus the next 2 nodes on the circle. When a node joins, the other nodes should integrate its distribution into their circles in a 'syncing' (not readable) state and automatically copy the appropriate keys and values into the new node, calculating the new two recovery nodes and causing them to check that the appropriate keys and values exist, and sending delete notifications for each key to all other nodes. In detail:

## Completed

* Command Line flags
* TLS Layer Prototype
* GOB streaming between servers
* Heartbeats
* Basic Memcache interface

## TODO

* Peer Swapping
* Autoconnect to all available nodes
* Proxying GOBs, Data
* Integrate consistenthash
* Disk-based key/value store 
* Expose Key/value store