# Database operation contract

Cacao keeps the existing SQLite schema. Existing files are inspected at startup;
there is no schema migration, added column/index/table, migration-version record,
data rewrite or automatic duplicate cleanup. Empty databases are initialized with
the original application tables. An incomplete existing schema produces an error
rather than an automatic alteration.

Storage initialization, schema validation and runtime startup run explicitly in
that order. Importing a Go package does not open a database or start background jobs.
Normal business writes still change records and may create SQLite's normal journal
files. This change does not enable WAL or change the persistent journal mode.

## Connections and transactions

- One pooled connection per process serializes local access. Reopened physical
  connections receive the same busy timeout and transaction settings.
- Business writes use `storage.Transaction`: `BEGIN IMMEDIATE`, at most eight
  attempts and a five-second deadline (or the caller's earlier cancellation).
  Each SQLite lock wait is 100 ms; retries back off from 10 to 320 ms. Only typed
  `SQLITE_BUSY` errors are retried. Validation, constraints and I/O failures are
  returned without replaying them.
- The wrapper owns the physical connection until commit or rollback completes.
  This includes commit contention: a failed commit must not leave an open SQLite
  transaction in the pool. A connection whose rollback fails is discarded.
- Callbacks use **only the supplied `tx`**. They must not call global model helpers,
  start another transaction, acquire application locks, write responses, change
  runtime maps or perform network I/O. Callbacks can run more than once; reset
  insert models/IDs on each attempt and expose results only after commit.
- Database functions return errors. Missing records are `gorm.ErrRecordNotFound`;
  successful empty lists are not errors. Configuration defaults apply only to
  missing keys. API requests have a ten-second budget (five minutes for GeoDB
  updates), respect earlier cancellation and distinguish business errors from
  storage failures. Unexpected failures are logged server-side.

## Uniqueness without schema changes

Username and network-name checks run inside the same IMMEDIATE transaction as
creation/update. Configuration writes look up and update/create inside that
transaction. This prevents duplicate creation through these Cacao operations
without adding unique indexes, and soft-deleted names/keys remain reusable.

Legacy duplicates are kept intact. Operations that would have to choose an
ambiguous username, configuration key, network identity or device identity return
an error instead of silently choosing, merging or deleting records. Listing records
continues to work. Resolve existing ownership/identity conflicts deliberately.

These checks govern code using this transaction contract; they cannot constrain
another program that writes directly to SQLite without the same checks. Run **one
Cacao process per storage directory**. Runtime network maps, in-memory device
sessions and GeoDB files are not coordinated across processes.

## Business and runtime consistency

Updates specify their fields. Credentials use a conditional update against the old
password and token; activity updates touch only `updated_at`. Stale requests cannot
restore an invalidated token or overwrite a newer password.

Registration policy, first-admin selection, account creation, the default network
and closing initial registration share one transaction. User/network deletion
soft-deletes associated active networks, devices and routes atomically. Expiration
checks are repeated during deletion instead of trusting an earlier snapshot.

Network mutations are ordered with `candy.ConfigMu`. Database commits happen before
runtime-map changes, client closure and success responses. Replacing a network
retires its old runtime object, including pending unauthenticated connections.
Socket writes have bounded deadlines to avoid indefinitely blocking runtime changes.

Device session IDs are **in memory only** (`gorm:"-"`), guarded by a per-identity
operation lock acquired before entering SQLite. Activation publishes a session
only after commit. Delayed traffic, metadata and offline writes from a replaced
session are rejected before writing. No session field is stored in the database.
Traffic counters increase monotonically. Background flush takes snapshots and
releases runtime-network locks before accessing SQLite. Failed offline writes are
logged and queued for the next minute's flush; stale/deleted-device retries are
discarded. Startup clears online flags, as before. Traffic since the last successful
flush can still be lost on abrupt process failure; the retry queue is not durable.

GeoDB publication precedes timestamp persistence. A timestamp failure keeps the
published file usable and retains a pending write; a later update or startup
recovery saves its timestamp.

## Delivery and verification

No database conversion or restore is needed to deploy this change. Retain the old
executable/image for code rollback. Rolling back code restores the old access
behavior; it does not undo intervening normal business writes. Maintain normal
operational backups and never overwrite a live database.

Each test package uses temporary storage and closes it afterward. Tests never use
a repository or deployment `sqlite.db`. Coverage includes independent connections
competing for writes, commit contention and replay, cancellation, error/panic
rollback, concurrent registration/name/configuration checks, atomic business
failure, stale credentials, API responses/cookies, live socket traffic with
concurrent flush/reconnect, offline retry and GeoDB recovery with real SQLite errors.
A legacy-schema regression test verifies byte-for-byte that startup validation does
not modify an existing database, including one containing duplicate records.

Run `go test -race ./...`, `go vet ./...` and `make all`. Pull requests run the race
suite before release-target builds. Tests use local servers and synthetic GeoDB
fixtures and do not require internet access.
