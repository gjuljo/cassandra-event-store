## OVERVIEW

This project is a inspired by [Mathew McLoughlin's talk at NDC London 2017](https://www.youtube.com/watch?v=9a1PqwFrMP0) and the more recent [article by Victor Martínez](https://victoramartinez.com/posts/event-sourcing-in-go/) about CQRS and Event Sourcing implementation in Go.
The objective is to have a simple and well known example to experiment with a basic **Event Store** implemented using [Apache Cassandra](https://cassandra.apache.org/) and to evaluate pros and cons of such an implementation.

--------------------------------------------------------------------------------------------------------------------------------

## KEYSPACES AND REPLICAS

According to the topology of the Cassandra cluster, you might need to create the `eventstore` keyspace as follows:

- single node in single datacenter

```
CREATE KEYSPACE eventstore WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};
```

- multiple nodes in single datacenter

```
CREATE KEYSPACE eventstore WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 3};
```

- multiple nodes in multiple datacenters

```
CREATE KEYSPACE eventstore WITH REPLICATION = {'class':'NetworkTopologyStrategy','DC1':3,'DC2':3,'DC3':3};
```

--------------------------------------------------------------------------------------------------------------------------------

## TABLES AND MATERIALIZED VIEWS

The solution uses just one table and one materialized view to allow queries by domain aggregate id and by event type respectively. The table might need an additional colum when lightweight transactions are not involved.

As you can see later, we use **batch statements** to insert events in the table to get the outcome of the batch transitions, so we can apply two different strategies:

- use **lightweight transitions** to know whether the batch statement has been applied

- use an additional **marker** in event to check the originator of the inserted event

### EVENT TABLE WITH LIGHTWEIGHT TRANSACTIONS

The following is a possible implementation of a Cassandra table to store events independently of the actual type of the event and its payload.
By adopting the UUID of the domain aggregate as **partition key** of the table, all the events are stored in the same partition and this speeds up queries and updates. Domain aggregate `version` also needed in the **primary key** of the table to let multiple events coexist in the same partition and also the `savetime` is needed to order them in the materialized view that, instead, gathers all the events of the same `type` in the same partition.

To handle the **optimistic locking** mechanism, we need a **static** column that allows us to use conditional statements in the batch statements and discard any table change in case the aggregate has evolved since the expected version that the events we wanted to persist were meant for.

```
CREATE TABLE IF NOT EXISTS eventstore.events (
  id               UUID,                -- uuid of the domain aggregate that the event is related to
  version          int,                 -- version of the domain aggregate generated by that event
  type             int,                 -- type of event
  payload          text,                -- actual payload of the event, typically in a JSON format 
  savetime         timestamp,           -- save time of the event, actually needed to order events in the materialized view
  current_version  int STATIC,          -- current version of the domain aggregate, corresponding to the biggest version
  PRIMARY KEY (id, version, savetime)
);
```

This table is suitable to be used when, programmatically, you are using **lightweight transactions** (also known as CAS, Compare And Set), so that you can get a feedback whether the batch has been actually applied or not, and understand if an optimistic locking issue occurred.

### EVENT TABLE WITHOUT LIGHTWEIGHT TRANSACTIONS

If we don't use the **ligthweight transactions**, we need to find an alternative way to realize whether an **optimistic locking** issue occurred as we don't get any direct feedback from the API about the actual batch execution. To do so we need to introduce an additional colum to the table, i.e. `marker`, just to let the code insert a sort of unique identifier for that batch execution and, later, verify with an subsequent query whether the our last event has been actually written with that marker. 

You can see how this actually works in the later examples.

```
CREATE TABLE IF NOT EXISTS eventstore.events (
  id               UUID,                -- uuid of the domain aggregate that the event is related to
  version          int,                 -- version of the domain aggregate generated by that event
  type             int,                 -- type of event
  payload          text,                -- actual payload of the event, typically in a JSON format 
  savetime         timestamp,           -- save time of the event, actually needed to order events in the materialized view
  marker           timeuuid,            -- marker needed to identify the actual writer of the event for optimistic locking check
  current_version  int STATIC,          -- current version of the domain aggregate, corresponding to the biggest version
  PRIMARY KEY (id, version, savetime)
);
```

### EVENT-BY-TYPE MATERIALIZED VIEW

This materialized view is meant to gather all the events of the same type in the same partition, to let queries such as "events by type". It enables processes that need to react whether a given type of event occurred.

```
CREATE MATERIALIZED VIEW eventstore.events_by_type AS
  SELECT id, version, type, payload, savetime FROM eventstore.events WHERE type IS NOT NULL AND version IS NOT NULL AND savetime IS NOT NULL
PRIMARY KEY (type, savetime, version, id);
```

--------------------------------------------------------------------------------------------------------------------------------

## CQL BATCH STATEMENTS

To implement the **optimistic locking strategy** we have to insert new events in the table only if the domain aggregate has not changed version since our previous read.

In Cassandra this is possible by writing all the events in a **batch statement**, using the `current_version` of the aggregate as condition.

### WITH LIGHTWEIGHT TRANSACTIONS

The following CQL statements are meant for the scenario when, programmatically, we use the **lightweight transaction** feature of Cassandra, so we have an immediate feedback from the API about the actual application of the batch statement.

When we add an event to the table we distinguish between two different cases:

- when the domain aggregate is new and it's not expected to be in the table

- when the domain aggregate is already existing and expected to be at a given `current_version`

#### FIRST EVENT

Here follows an example for a batch statement meant to insert one event for a new domain aggregate in the table:

```
BEGIN BATCH
INSERT INTO eventstore.events (id, current_version) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 1) IF NOT EXISTS;
INSERT INTO eventstore.events (id, version, type, payload, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 1, 11, 'aaa', toTimeStamp(now()));
APPLY BATCH;
```

If the batch is successfully applied, the table is supposed to contain the following row:

#### *events* table

| id (P) | version (C) | savetime (C) | current_version (S) | payload | type |
|--------|------------:|--------------|--------------------:|--------:|-----:| 
|fade87a1-9df9-46bb-aae6-63b2b763094d | 1 | 2020-11-24 18:21:49.826000+0000 | 1 | 'aaa' | 11 |

Correspondingly, the materialized view is supposed to have the following contents:

#### *events_by_type* view

| type (P) | savetype (C) | version (C) | id | payload |
|----------|--------------|------------:|----|--------:| 
|11| 2020-11-24 18:21:49.826000+0000 | 1 | fade87a1-9df9-46bb-aae6-63b2b763094d|'aaa'|

#### FURTHER EVENTS

The next batch statement shows how to insert subsequent events when the domain aggregate is supposed to already exist with an expected `current_version` (in this case `1`):

```
BEGIN BATCH
UPDATE eventstore.events SET current_version = 3 WHERE id = fade87a1-9df9-46bb-aae6-63b2b763094d IF current_version = 1;
INSERT INTO eventstore.events (id, version, type, payload, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 2, 22, 'bbb', toTimeStamp(now()));
INSERT INTO eventstore.events (id, version, type, payload, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 3, 33, 'ccc', toTimeStamp(now()));
APPLY BATCH;
```

If the batch is successfully applied, the table is supposed to contain the following rows:

#### `events` table

| id (P) | version (C) | savetime (C) | current_version (S) | payload | type |
|--------|------------:|--------------|--------------------:|--------:|-----:|
|fade87a1-9df9-46bb-aae6-63b2b763094d | 1 | 2020-11-24 18:21:49.826000+0000 | 3 | 'aaa' | 11 |
|fade87a1-9df9-46bb-aae6-63b2b763094d | 2 | 2020-11-24 18:21:49.827000+0000 | 3 | 'bbb' | 22 |
|fade87a1-9df9-46bb-aae6-63b2b763094d | 3 | 2020-11-24 18:21:49.828000+0000 | 3 | 'ccc' | 33 |

Please notice the `current_version` column that, being static, has the same value for all the rows and it is equal to the latest version of the domain aggregate.

The corresponding materialized view should be the following:

#### `events_by_type` view

| type (P) | savetype (C) | version (C) | id | payload |
|----------|--------------|------------:|----|--------:|
|11| 2020-11-24 18:21:49.826000+0000 | 1 | fade87a1-9df9-46bb-aae6-63b2b763094d|'aaa'|
|22| 2020-11-24 18:21:49.827000+0000 | 2 | fade87a1-9df9-46bb-aae6-63b2b763094d|'bbb'|
|33| 2020-11-24 18:21:49.828000+0000 | 3 | fade87a1-9df9-46bb-aae6-63b2b763094d|'ccc'|

### WITHOUT LIGHTWEIGHT TRANSACTIONS

As we said before, if we don't leverage the **ligthweight transaction** feature of the batch update we need to introduce ad **additional column** in the event record, i.e. the `marker`. This column is needed to check whether the batch statement has been actually applied or not.

#### FIRST EVENT

The following statement is like the one discussed before, with the addition of the extra column for the `marker` (typically a time UUID).

```
BEGIN BATCH
INSERT INTO eventstore.events (id, current_version) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 1) IF NOT EXISTS;
INSERT INTO eventstore.events (id, version, type, payload, marker, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 1, 11, 'aaa', 58f660ca-2f2d-11eb-a263-00155d875616, toTimeStamp(now()));
APPLY BATCH;
```

When the batch statement is successfully applied, the table should include the following row:

#### `events` table

| id (P) | version (C) | savetime (C) | current_version (S) | marker | payload | type |
|--------|------------:|--------------|--------------------:|--------|--------:|-----:|
|fade87a1-9df9-46bb-aae6-63b2b763094d | 1 | 2020-11-24 18:21:49.826000+0000 | 1 | 58f660ca-2f2d-11eb-a263-00155d875616 | 'aaa' | 11 |

Programmatically, assuming we are not getting any feedback from the API as we are not using lightweight transactions, we need to run the following query to fetch the actual value of the `marker` and check whether it matches with the expected value (i.e. `58f660ca-2f2d-11eb-a263-00155d875616`).
If it doesn't match, then it means that someone else as written the event before us and we had an **optimistic locking** collision.

```
SELECT marker FROM events WHERE id = fade87a1-9df9-46bb-aae6-63b2b763094d AND version = 1
```

In a multi datacenter environment, if we assume that the batch statement is executed in a `QUORUM` **consistency level** and this subsequent query in a `LOCAL_QUORUM` level, the overall performance of this approach should be better than using **lightweight transactions**.

#### `events_by_type` view

| type (P) | savetype (C) | version (C) | id | payload |
|----------|--------------|------------:|----|--------:|
|11| 2020-11-24 18:21:49.826000+0000 | 1 | fade87a1-9df9-46bb-aae6-63b2b763094d|'aaa'|

#### FURTHER EVENTS

Here follows the batch statement to insert subsequent events for an expected version of a domain aggregate.

```
BEGIN BATCH
UPDATE eventstore.events SET current_version = 3 WHERE id = fade87a1-9df9-46bb-aae6-63b2b763094d IF current_version = 1;
INSERT INTO eventstore.events (id, version, type, payload, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 2, 22, 'bbb', 98ea964c-2f2d-11eb-b476-00155d875616, toTimeStamp(now()));
INSERT INTO eventstore.events (id, version, type, payload, savetime) VALUES (fade87a1-9df9-46bb-aae6-63b2b763094d, 3, 33, 'ccc', 98ea964c-2f2d-11eb-b476-00155d875616, toTimeStamp(now()));
APPLY BATCH;
```

This is the resulting table:

#### `events` table

| id (P) | version (C) | savetime (C) | current_version (S) | marker | payload | type |
|--------|------------:|--------------|--------------------:|--------|--------:|-----:|
|fade87a1-9df9-46bb-aae6-63b2b763094d | 1 | 2020-11-24 18:21:49.826000+0000 | 3 | 98ea964c-2f2d-11eb-b476-00155d875616| 3 | 'aaa' | 11 |
|fade87a1-9df9-46bb-aae6-63b2b763094d | 2 | 2020-11-24 18:21:49.827000+0000 | 3 | 98ea964c-2f2d-11eb-b476-00155d875616| 3 | 'bbb' | 22 |
|fade87a1-9df9-46bb-aae6-63b2b763094d | 3 | 2020-11-24 18:21:49.828000+0000 | 3 | 98ea964c-2f2d-11eb-b476-00155d875616| 3 | 'ccc' | 33 |

And this is the corresponding materialized view:

#### `events_by_type` view

| type (P) | savetype (C) | version (C) | id | payload |
|----------|--------------|------------:|----|--------:| 
|11| 2020-11-24 18:21:49.826000+0000 | 1 | fade87a1-9df9-46bb-aae6-63b2b763094d|'aaa'|
|22| 2020-11-24 18:21:49.827000+0000 | 2 | fade87a1-9df9-46bb-aae6-63b2b763094d|'bbb'|
|33| 2020-11-24 18:21:49.828000+0000 | 3 | fade87a1-9df9-46bb-aae6-63b2b763094d|'ccc'|

In this case, to check the **optimistic locking**, we should run the following query:

```
SELECT marker FROM events WHERE id = fade87a1-9df9-46bb-aae6-63b2b763094d AND version = 3
```

Please notice that we are checking the `version` of the latest row we expect to have written and we shouldn't care the `current_version` of the domain aggregate as we should not exclude that it has evolved in between our batch statement and the subsequent select.

--------------------------------------------------------------------------------------------------------------------------------

## QUERIES AT HAND

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
- [PatientMangement project on GitHub - Mathew McLoughlin](https://github.com/mat-mcloughlin/PatientMangement)
- [Event Sourcing in Go - Victor Martínez](https://victoramartinez.com/posts/event-sourcing-in-go/)
- [Building Microservices with Event Sourcing/CQRS in Go using gRPC, NATS Streaming and CockroachDB - Shiju Varghese](https://shijuvar.medium.com/building-microservices-with-event-sourcing-cqrs-in-go-using-grpc-nats-streaming-and-cockroachdb-983f650452aa)
- [Simple CQRS Implementation in C# - Gergory Young](https://github.com/gregoryyoung/m-r/tree/master/SimpleCQRS)
