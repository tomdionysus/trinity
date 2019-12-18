![Trinity DB Logo](../gfx/trinity_m.png) 

# Trinity DB - Consistency Modes

Trinity is intended to have 3 consistency modes, selectable per client session.

## WRITE_UNCOMMITTED

Writes will return immediately - the write is not guaranteed to have been persisted to disk on any node node, the disk persistence and replication on nodes will happen asynchronously. This is the fastest and least consistent mode.

## WRITE_COMMITED

Writes will return when they have been persisted to disk on at least one node, the replication to other nodes will happen asynchronously. This is the **default** mode and represents a balance between performance and consistency.

## WRITE_REPLICATED

Writes will return only after they have been persisted on all replication nodes - the write is guaranteed to have been persisted across multiple nodes. This is the slowest and most consistent mode.
