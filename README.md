# Linkforge

A URL shortener with real-time click analytics, built incrementally as the course project for **Golang for developers**.

The design is specified in the project TDD; this repository is the skeleton it describes.

## Current state

**Milestones M1–M2 and M4 are implemented.** Linkforge includes a Base62 codec, a domain Shortener service, an in-memory store, and reusable LRU cache, trie, TopK heap, and short-code generator. M3 functionality is partially implemented. The remaining milestones are still under development.


Each file declares exactly one interface alongside its no-op implementation, so the compile-time contract and its placeholder stay together. Types shared by several interfaces live in the package's `model.go`.

## Layout

```
cmd/      linkforge         service binary — prints its version, nothing more
          linkctl           operator CLI   — prints its version, nothing more

pkg/                        reusable libraries — a separate Go module
  cache/    LRU[K,V]        bounded least-recently-used cache         M4
  trie/     Trie            prefix index for alias autocomplete       M4
  topk/     TopK            top-N ranking using a bounded heap        M4
  shortid/  CodeGenerator  random Base62 short-code generator         M4


internal/                   application code
  base62/                   base62 encode/decode                       M1
  store/                    in-memory link store                       M1
  link/     Store           link persistence, declared by its consumer M1 / M7
            Shortener       the domain service behind every transport  M2
  click/    Ingester        non-blocking submission from the hot path  M5
            Store           append-only raw event persistence          M7
            Aggregator      time series and top-N queries              M5
            Publisher       fan-out to live WebSocket subscribers      M9
  auth/     Issuer          access token minting                       M9
            Verifier        token validation for middleware            M9
            UserStore       account persistence                        M7
  transport/Server          uniform listener lifecycle                 M6 / M8
  config/   Loader          environment-only configuration             M3
  observability/Metrics     counters and histograms                    M10
```

The `pkg/` tree is a second module so that reusable libraries stay free of any dependency on application code; `go.work` ties the two together for local development.

## Design rules

These hold for every milestone.

1. **Interfaces are declared by the consumer.** `link.Store` lives in `internal/link` because the shortener is what needs it — not beside any implementation. Constructors return concrete types.
2. **The dependency arrow points inwards.** `internal/link` and `internal/click` may not import a transport or a storage package.
3. **Analytics never blocks a redirect.** `Ingester.Submit` drops events under saturation rather than applying backpressure.
4. **Every listener shares one lifecycle contract**, which is what makes a single coordinated graceful shutdown possible.

## Working on it

```bash
make            # vet, lint and test — the same gates CI runs
make build
make test
make cover
make bench
make tidy
```

Requires Go 1.25+. Linting uses [golangci-lint](https://golangci-lint.run) (`brew install golangci-lint`).

## Milestones

| # | Lessons | Scope |
|---|---|---|
| M0 | 1 | Toolchain, module layout, CI |
| M1 | 2–3 | base62 codec, `Link` model, in-memory store |
| M2 | 4–5 | Methods, interfaces, generics, error values |
| M3 | 6–7 | Files and `context`, CLI, `slog`, tests |
| M4 | 8–9 | Own hash table, LRU, trie, top-N heap |
| M5 | 10–11 | Ingestion pipeline, worker pool, synchronisation |
| M6 | 12–13 | TCP resolver, graceful shutdown, Docker |
| M7 | 14 | PostgreSQL and MongoDB stores |
| M8 | 15–16 | `net/http` first, then a framework; REST and Swagger |
| M9 | 17–19 | Validation, JWT and middleware, WebSockets |
| M10 | 20–21 | Production hardening and project defence |


## Shortener Service

The Shortener service provides four operations:

- `Shorten` — creates a short link with a generated code or custom alias.
- `Resolve` — retrieves a link by code and checks its expiration.
- `List` — lists links by owner with cursor-based pagination.
- `Delete` — removes a link after verifying ownership.

The service uses the `link.Store` interface and an in-memory implementation.

## Testing

Run all tests:

```bash
go test ./...
```

## Complexity

| Structure | Operation | Average | Worst case |
|---|---|---|---|
| LRU | Get | O(1) | O(n) |
| LRU | Put | O(1) | O(n) |
| Trie | Insert | O(L) | O(L) |
| Trie | WithPrefix | O(L + k·W) | O(L + 256·m + S) |
| TopK | Add | O(log N) | O(log N) |
| TopK | Result | O(N log N) | O(N log N) |

Here, n is the number of stored entries, L is the length of a word or prefix, k is the number of returned words, W is the average length of a returned word's suffix, S is the total work required to traverse the matching subtree, and N is the heap capacity.

LRU.Get remains O(1) on average even though it updates the usage order. A hash map locates the entry, and a doubly linked list moves it to the front in constant time without scanning other entries.


## Benchmarks

Benchmarks were run on Windows (amd64) with an AMD Ryzen 7 5800X 8-Core Processor.

Command:

```bash
make bench
```

Results:

```text
BenchmarkLRUGet-16            122433163       9.752 ns/op       0 B/op       0 allocs/op
BenchmarkTrieWithPrefix-16       106575       11287 ns/op     664 B/op      16 allocs/op
```

The LRU cache hit requires no heap allocations and takes about 9.8 ns per operation. Trie prefix search takes about 11.3 microseconds and allocates memory to build the returned strings and slice.

The Trie benchmark uses approximately 10,000 words and requests up to 10 matching results.

