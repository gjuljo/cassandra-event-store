


# 1. CASSANDRA KEYSPACE AND TABLES

### KEYSPACES AND REPLICAS
```
CREATE KEYSPACE eventstore WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};
```

```
CREATE KEYSPACE eventstore WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 3};
```

```
CREATE KEYSPACE eventstore WITH REPLICATION = {'class':'NetworkTopologyStrategy','DC1':3,'DC2':3,'DC3':3};
```

### EVENT TABLE
```
CREATE TABLE IF NOT EXISTS eventstore.events (
  id           UUID,
  version      int,
  type         int,
  payload      text,
  savetime     timestamp,
  current_version  int STATIC,
  PRIMARY KEY (id, version, savetime)
);
```

### EVENT-BY-TYPE MATERIALIZED VIEW
```
CREATE MATERIALIZED VIEW eventstore.events_by_type AS
  SELECT id, version, type, payload, savetime FROM eventstore.events WHERE type IS NOT NULL AND version IS NOT NULL AND savetime IS NOT NULL
PRIMARY KEY (type, savetime, version, id);
```

# 2. CQL BATCH QUERIES
```
BEGIN BATCH
INSERT INTO eventstore.events (id, current_version) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 1) IF NOT EXISTS;
INSERT INTO eventstore.events (id, version, type, payload, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 1, 11, 'aaa', toTimeStamp(now()));
APPLY BATCH;
```

## *events*

| `id (P)` | `version (C)` | `savetime (C)` | `current_version (S)` | `payload` | `type` |
|----------|--------------:|----------------|----------------------:|----------:|-------:| 
|`fade87a1-9df9-46bb-aae6-63b2b763094d` | `1` | `2020-11-24 18:21:49.826000+0000` | `1` | `'aaa'` | `11` |


## *events_by_type*

| `type (P)` | `savetype (C)` | `version (C)` | `id` | `payload` |
|------------|----------------|--------------:|------|----------:| 
|`11`| `2020-11-24 18:21:49.826000+0000` | `1` | `fade87a1-9df9-46bb-aae6-63b2b763094d`|`'aaa'`|




if expectedVersion > 0:

```
BEGIN BATCH
UPDATE eventstore.events SET current_version = 3 WHERE id = fade87a1-9df9-46bb-aae6-63b2b763094d IF current_version = 1;
INSERT INTO eventstore.events (id, version, type, payload, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 2, 22, 'bbb', toTimeStamp(now()));
INSERT INTO eventstore.events (id, version, type, payload, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 3, 33, 'ccc', toTimeStamp(now()));
APPLY BATCH;
```

## *events*

| `id (P)` | `version (C)` | `savetime (C)` | `current_version (S)` | `payload` | `type` |
|----------|--------------:|----------------|----------------------:|----------:|-------:| 
|`fade87a1-9df9-46bb-aae6-63b2b763094d` | `1` | `2020-11-24 18:21:49.826000+0000` | `3` | `'aaa'` | `11` |
|`fade87a1-9df9-46bb-aae6-63b2b763094d` | `2` | `2020-11-24 18:21:49.827000+0000` | `3` | `'bbb'` | `22` |
|`fade87a1-9df9-46bb-aae6-63b2b763094d` | `3` | `2020-11-24 18:21:49.828000+0000` | `3` | `'ccc'` | `33` |

## *events_by_type*

| `type (P)` | `savetype (C)` | `version (C)` | `id` | `payload` |
|------------|----------------|--------------:|------|----------:| 
|`11`| `2020-11-24 18:21:49.826000+0000` | `1` | `fade87a1-9df9-46bb-aae6-63b2b763094d`|`'aaa'`|
|`22`| `2020-11-24 18:21:49.827000+0000` | `2` | `fade87a1-9df9-46bb-aae6-63b2b763094d`|`'bbb'`|
|`33`| `2020-11-24 18:21:49.828000+0000` | `3` | `fade87a1-9df9-46bb-aae6-63b2b763094d`|`'ccc'`|



### HANDY QUERIES
```
SELECT * FROM eventstore.events LIMIT 100;
SELECT * FROM eventstore.events_by_type LIMIT 100;

SELECT COUNT(*) FROM eventstore.events_by_type WHERE type = 2;
SELECT * FROM eventstore.events_by_type WHERE type = 1 AND savetime >= 1606154195500;
SELECT * FROM eventstore.events_by_type WHERE type = 1 AND savetime >= '2020-11-23';

TRUNCATE TABLE eventstore.events;
DROP TABLE eventstore.events;
DROP MATERIALIZED VIEW eventstore.events_by_type;
```

# 3. References
- [An Introduction to CQRS and Event Sourcing Patterns - Mathew McLoughlin]([https://www.youtube.com/watch?v=9a1PqwFrMP0])
- [Event Sourcing in Go - Victor Martínez](https://victoramartinez.com/posts/event-sourcing-in-go/)
- [Building Microservices with Event Sourcing/CQRS in Go using gRPC, NATS Streaming and CockroachDB - Shiju Varghese](https://shijuvar.medium.com/building-microservices-with-event-sourcing-cqrs-in-go-using-grpc-nats-streaming-and-cockroachdb-983f650452aa)
- [Simple CQRS Implementation in C# - Gergory Young](https://github.com/gregoryyoung/m-r/tree/master/SimpleCQRS)
