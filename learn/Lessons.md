# Go in 3 Days: Practical Exercises for .NET Developers

A hands-on exercise plan for a senior .NET developer. Each day is about 6 to 8 hours: read a little, write a lot. Don't read a Go book first. Do the exercises, and look things up in the Tour of Go and pkg.go.dev when stuck.

## Setup (30 minutes)

1. Install Go from go.dev/dl, then check `go version`.
2. Editor: VS Code with the official Go extension, or GoLand. Accept the prompts to install `gopls`, `dlv` and friends.
3. Make a workspace: `mkdir gobootcamp && cd gobootcamp && go mod init example.com/gobootcamp`
4. Run the interactive Tour of Go (go.dev/tour) sections Basics, Methods and Interfaces, and Concurrency at double speed. It takes about 60 to 90 minutes. Skim, don't linger.
5. Skim Effective Go (go.dev/doc/effective\_go) once. You will return to it.

## The .NET to Go mental map

| .NET | Go | Notes |
| --- | --- | --- |
| Solution / csproj / NuGet | Module (`go.mod`) / package / `go get` | One module per repo is typical. No separate project files. |
| Namespace | Package (one per directory) | Exported = Capitalized identifier. No `public` keyword. |
| Class | `struct` + methods | No inheritance. Use composition (embedding). |
| Interface (explicit) | Interface (implicit) | A type satisfies an interface just by having the methods. Define interfaces where they are consumed, and keep them small. |
| Exceptions | `error` return values | `if err != nil`. `panic` is for truly unrecoverable bugs only. |
| `try/finally`, `using` | `defer` | Runs at function exit, LIFO. |
| `List<T>`, `Dictionary<K,V>` | slice `[]T`, `map[K]V` | Slices are views over arrays. Learn append semantics. |
| `null` | `nil` (pointers, slices, maps, interfaces, funcs, channels) | Strings and structs are never nil. Zero values are useful by design. |
| Generics | Generics (since 1.18) | Simpler than C#. Use sparingly. |
| `Task` / `async` / `await` | goroutines + channels | No colored functions. Blocking code is normal. |
| `CancellationToken` | `context.Context` | First parameter of anything that blocks or does I/O. |
| xUnit / NUnit | `testing` package, `go test` | Table-driven tests are the norm. |
| ASP.NET Core | `net/http` (+ optional router) | The standard library is production grade. |
| EF Core / Dapper | `database/sql` (+ sqlc, pgx, sqlx) | Closer to Dapper than EF. |
| `System.Text.Json` | `encoding/json` + struct tags |  |
| ILogger | `log/slog` | Structured logging in the stdlib. |
| LINQ | `slices`, `maps` packages, plain loops | Write the `for` loop. That is idiomatic. |
| Nullable value types | zero values, pointers, `ok` idiom |  |

Three habits to unlearn early:

- Don't build deep abstractions, DI containers or class hierarchies. Pass dependencies via constructors (plain functions) and keep things flat.
- Don't name things `IFoo`, `FooService`, `FooManager`. Prefer short names and small packages.
- Don't fight `if err != nil`. Handle errors at each step and add context with `fmt.Errorf("doing X: %w", err)`.

Always run these as you go: `go fmt ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`.

---

# Day 1: Language fundamentals

Goal: be fluent with syntax, types, slices, maps, structs, methods, errors and `defer`. Each exercise is a small CLI program in its own folder under `day1/`.

## Exercise 1.1: Word frequency counter (60 min)

Read a text file (path from `os.Args` or the `flag` package), and print the top N most frequent words.

- Use `bufio.Scanner` and `strings.Fields`.
- Use a `map[string]int`, then sort by count using `sort.Slice` or `slices.SortFunc`.
- Normalize case and strip punctuation (`unicode.IsLetter`).
- Handle file-not-found with a clean error message and a non-zero exit code.

Learn: packages, `main`, `flag`, maps, slices, sorting, error returns, `defer file.Close()`.

Stretch: add a `-n` flag for top-N and a `-ignore` flag for a comma-separated stop word list.

## Exercise 1.2: Slice semantics lab (45 min)

Write a program that demonstrates, with printed output, each of these:

- Length vs capacity, and what `append` does when capacity is exceeded.
- Two slices sharing a backing array, and a mutation through one visible in the other.
- The classic bug: appending to a sub-slice overwrites the parent's data.
- Fixing it with a full slice expression `s[a:b:b]` or `slices.Clone`.
- Passing a slice to a function and mutating vs appending inside it.
- Iterating a map twice and noticing the order differs (map order is intentionally random).

Learn: this is the single biggest source of surprise for newcomers. Do not skip it.

## Exercise 1.3: Shapes with interfaces (45 min)

Define a `Shape` interface with `Area() float64` and `Perimeter() float64`. Implement `Circle`, `Rect` and `Triangle` as structs.

- Write `func Largest(shapes []Shape) Shape`.
- Implement `String() string` on each type so `fmt.Println` works nicely (that is the `Stringer` interface).
- Use a type switch to print type-specific details.
- Use a value receiver on one type and a pointer receiver on another. Then try to put each into a `[]Shape` and read the compiler errors. Work out the method-set rules.

Learn: implicit interfaces, receivers (value vs pointer), type switches, `Stringer`.

Stretch: add an `io.Writer`-accepting function `Describe(w io.Writer, s Shape)` and call it with `os.Stdout` and a `bytes.Buffer`.

## Exercise 1.4: Errors done properly (60 min)

Build a small bank-account package `bank` with `Deposit`, `Withdraw` and `Transfer`.

- Define a sentinel error `ErrInsufficientFunds`.
- Define a custom error type `*TxError` carrying account ID and amount.
- Wrap errors with `fmt.Errorf("transfer %s to %s: %w", ...)`.
- In `main`, use `errors.Is` and `errors.As` to branch on error kinds.
- Use `defer` to make `Transfer` roll back on failure.
- Write a function that recovers from a `panic` and converts it to an error, so you understand `recover` (then never use it casually).

Learn: error wrapping, `errors.Is/As`, custom errors, `defer`/`recover`.

## Exercise 1.5: Generic collection helpers (45 min)

Write generic functions: `Map[T, U any]`, `Filter[T any]`, `Reduce[T, A any]`, and a `Set[T comparable]` type with `Add`, `Has`, `Union`.

- Compare your versions with the standard `slices` and `maps` packages.
- Notice what Go generics cannot do (no generic methods on non-generic types, no overloads).

Learn: type parameters, constraints, `comparable`, `cmp.Ordered`.

## Exercise 1.6: JSON round trip (45 min)

Define structs for a `Customer` with nested `Address` and a list of `Order`s. Use struct tags (`json:"first_name,omitempty"`).

- Marshal to indented JSON and unmarshal it back.
- Handle a `time.Time` field and an optional field via a pointer.
- Write a custom `UnmarshalJSON` for an enum-like status type.
- Decode an unknown payload into `map[string]any` and walk it.
- Use `json.NewDecoder(...).DisallowUnknownFields()` to see strict parsing.

Learn: tags, `omitempty`, pointers for optionals, custom marshalers.

## Day 1 checkpoint

You should be able to explain, without looking: zero values, slice vs array, value vs pointer receivers, why `for i, v := range` copies `v`, how `defer` order works, and when to return `(T, error)`.

---

# Day 2: Concurrency, testing and the standard library

Goal: goroutines, channels, `sync`, `context`, table-driven tests, and the stdlib I/O model.

## Exercise 2.1: Goroutines and WaitGroup (45 min)

Write a program that launches 10 goroutines, each sleeping a random time then printing its ID.

- Use `sync.WaitGroup` to wait.
- Deliberately create a data race: 100 goroutines incrementing a shared counter. Run with `go run -race` and read the report.
- Fix it with `sync.Mutex`, then again with `sync/atomic`.
- Note the loop variable capture behavior (changed in Go 1.22: each iteration gets its own variable).

Learn: `go` keyword, `WaitGroup`, races, mutex, atomics, the race detector.

## Exercise 2.2: Pipeline with channels (60 min)

Build a three-stage pipeline: generator (emits ints) -> squarer -> printer, connected by channels.

- Use directional channel types (`<-chan int`, `chan<- int`).
- Close channels from the sender side and use `for v := range ch`.
- Make the squarer stage run N workers in parallel, then fan the results back in to one channel.
- Observe what happens with unbuffered vs buffered channels. Cause a deadlock on purpose and read the runtime message.

Learn: channel semantics, closing, fan-out/fan-in, deadlocks.

## Exercise 2.3: Concurrent URL checker with cancellation (75 min)

Take a list of URLs (file or args) and check them concurrently, printing status code and latency.

- Limit concurrency to K using a buffered channel as a semaphore, or `golang.org/x/sync/errgroup` with `SetLimit`.
- Use `http.NewRequestWithContext` and a per-request timeout via `context.WithTimeout`.
- Cancel everything on Ctrl+C using `signal.NotifyContext`.
- Collect results into a slice safely (index per goroutine, or via channel).
- Use `select` to handle result vs `ctx.Done()`.

Learn: `context`, `select`, `errgroup`, bounded concurrency, graceful shutdown. This is your `Task.WhenAll` + `CancellationToken` equivalent.

## Exercise 2.4: Rate-limited worker pool (60 min)

Implement a worker pool: a jobs channel, W workers, a results channel.

- Jobs simulate work with `time.Sleep` and sometimes return an error.
- Add a `time.Ticker` to rate-limit to 5 jobs per second.
- Support graceful shutdown: stop accepting jobs, let in-flight ones finish, then exit.
- Add a timeout so shutdown does not hang forever.

Learn: worker pool pattern, tickers, shutdown sequencing, goroutine leaks.

Stretch: use `goleak` (go.uber.org/goleak) in a test to prove nothing leaks.

## Exercise 2.5: Thread-safe cache (45 min)

Build a generic `Cache[K comparable, V any]` with `Get`, `Set` and TTL expiry.

- Protect it with `sync.RWMutex`.
- Add a background goroutine that evicts expired entries and stops when `ctx` is cancelled (or on `Close()`).
- Benchmark vs `sync.Map` using `testing.B`.

Learn: `RWMutex`, generics plus concurrency, benchmarks.

## Exercise 2.6: Table-driven tests (60 min)

Write tests for earlier exercises (the bank package and your generic helpers).

- Use the table-driven pattern with `t.Run(tc.name, ...)` subtests.
- Use `t.Helper()`, `t.Cleanup()`, and `t.TempDir()`.
- Add a fake for an interface dependency (write it by hand, no mocking framework).
- Run `go test -cover ./...` and open the HTML report with `go tool cover -html=cover.out`.
- Add one fuzz test (`func FuzzX(f *testing.F)`) for a parsing function.
- Run `go test -race ./...` on everything.

Learn: `testing`, subtests, fakes, coverage, fuzzing.

## Exercise 2.7: Streams with io.Reader and io.Writer (30 min)

Write a tool that reads from stdin or a file, wraps the reader with a gzip reader or an `io.TeeReader` that counts bytes, and writes to stdout or a file.

- Compose with `io.Copy`, `io.MultiWriter`, `bufio`.
- Notice how everything composes because the interfaces are tiny.

Learn: `io.Reader`/`io.Writer` composition, the design philosophy of Go's stdlib.

## Day 2 checkpoint

You should be able to explain: when to use a channel vs a mutex, why you close channels from the sender, how `context` propagates cancellation, and what a goroutine leak looks like.

---

# Day 3: Build a real service

Goal: assemble everything into a small production-shaped HTTP service with persistence, tests and a container image.

## Exercise 3.1: Minimal HTTP API with net/http (75 min)

Build a `tasks` service (a todo list) using only the standard library (Go 1.22+ has method and path patterns in `http.ServeMux`).

- Routes: `GET /tasks`, `POST /tasks`, `GET /tasks/{id}`, `PUT /tasks/{id}`, `DELETE /tasks/{id}`.
- Use an in-memory store behind an interface (`TaskStore`), guarded by a mutex.
- Parse path values with `r.PathValue("id")`. Return correct status codes (201, 404, 400, 204).
- Write helper functions `writeJSON` and `readJSON`, and a consistent error response shape.
- Use `http.Server` with explicit `ReadTimeout`, `WriteTimeout` and `IdleTimeout`.

Learn: handlers, `ServeMux`, request/response, JSON helpers, server config.

## Exercise 3.2: Middleware (45 min)

Implement middleware as `func(http.Handler) http.Handler`.

- Request logging with `log/slog` (method, path, status, duration). You will need a small `ResponseWriter` wrapper to capture status.
- Panic recovery middleware returning 500.
- Request ID middleware that stores an ID in `context` and includes it in logs.
- A simple API key auth middleware.
- Chain them with a small `Chain(h, mws...)` helper.

Learn: the Go equivalent of ASP.NET middleware, context values (use sparingly), wrapping `ResponseWriter`.

## Exercise 3.3: Graceful shutdown (30 min)

Wire `signal.NotifyContext` to `server.Shutdown(ctx)` with a 10 second deadline. Verify that in-flight requests complete when you Ctrl+C.

## Exercise 3.4: Real persistence (90 min)

Replace the in-memory store with SQL. Use SQLite (pure Go driver `modernc.org/sqlite`, no cgo) or Postgres via Docker with `github.com/jackc/pgx/v5`.

- Use `database/sql`: `db.QueryContext`, `rows.Scan`, `rows.Close()`, `rows.Err()`.
- Always use parameterized queries and pass `r.Context()` through.
- Write a simple migration step (a `schema.sql` run at startup, or `golang-migrate`).
- Implement transactions with `BeginTx`, `defer tx.Rollback()`, `tx.Commit()`.
- Read configuration from environment variables into a `Config` struct (no DI container, just `main` wiring everything).

Learn: `database/sql` ergonomics, connection pooling settings, context propagation, transactions.

Stretch: try `sqlc` to generate type-safe code from SQL and compare it with hand-written queries.

## Exercise 3.5: Test the API (60 min)

- Use `net/http/httptest` to test handlers without a running server.
- Test the store against a real SQLite in-memory or temp-file database.
- Test middleware in isolation.
- Add one end-to-end test using `httptest.NewServer`.

Learn: `httptest`, testing at several layers, no mocking framework needed.

## Exercise 3.6: Package layout and CLI (45 min)

Restructure the project:

```
cmd/tasksd/main.go      // wiring only
internal/store/         // SQL implementation
internal/api/           // handlers and middleware
internal/config/        // env parsing
```

- Understand `internal/` (compiler-enforced privacy).
- Add a second binary `cmd/taskctl` that calls your API with a flag-based CLI (try `cobra` if you want to see what most real CLIs use).
- Cross-compile: `GOOS=linux GOARCH=amd64 go build -o bin/tasksd ./cmd/tasksd`.

## Exercise 3.7: Containerize (30 min)

Write a multi-stage Dockerfile: build with the `golang` image, copy the static binary into `scratch` or `gcr.io/distroless/static`. Set `CGO_ENABLED=0`. Compare image size to a typical .NET image and enjoy.

## Exercise 3.8: Profile and tune (30 min, optional)

- Add `net/http/pprof` on a separate local port.
- Generate load (`hey` or `vegeta`), capture a CPU profile and open it with `go tool pprof -http=:8081`.
- Write a benchmark with `-benchmem` and look at allocations.

## Day 3 checkpoint

You should now be able to scaffold a new Go service from memory, explain the package layout, and write tests at the handler and store levels.

---

# Capstone options (pick one if you have time left)

1. Port a small .NET utility or library you know well to Go. Porting something familiar is the best way to feel the idioms.
2. Build a concurrent log tailer and aggregator: tail multiple files, parse lines, and expose counters on `/metrics` as JSON.
3. Build a URL shortener with SQLite, a redirect handler, click counts and rate limiting.
4. Build a small TCP chat server with `net`, one goroutine per connection and a broadcast hub goroutine.

# Gotchas that will bite a .NET developer

- A nil interface is not the same as an interface holding a nil pointer. Returning a nil `*MyErr` as `error` yields a non-nil error.
- `range` over a slice yields copies of elements. Mutate via the index (`s[i].X = 1`) or use a slice of pointers.
- Maps are not safe for concurrent writes. The runtime will crash your program.
- Writing to a nil map panics, but reading from one is fine.
- Unused variables and imports are compile errors. Embrace it.
- Strings are immutable UTF-8 byte sequences. `len(s)` is bytes, not characters. Use `[]rune` or `utf8.RuneCountInString`.
- Don't start a goroutine without knowing how it will stop.
- `http.DefaultClient` and `http.ListenAndServe` have no timeouts. Always set them.
- Closing a closed channel panics, and sending on a closed channel panics.
- `init()` functions and package-level state are legal but discouraged. Prefer explicit construction.

# Reference list

- Tour of Go: go.dev/tour
- Effective Go: go.dev/doc/effective\_go
- Go by Example: gobyexample.com (excellent for quick syntax lookups)
- Standard library: pkg.go.dev/std
- Go Code Review Comments: go.dev/wiki/CodeReviewComments
- Google Go Style Guide: google.github.io/styleguide/go
- 100 Go Mistakes and How to Avoid Them (book and site, 100go.co) for after the 3 days

# Suggested daily rhythm

- Morning (3 hours): 2 to 3 exercises from the day, writing from scratch with no AI autocomplete. The muscle memory matters.
- Afternoon (3 hours): remaining exercises, plus reading the standard library source for one package you used (try `bufio` or `net/http`).
- Evening (1 hour): run `go vet`, read the Code Review Comments page, refactor one earlier exercise to be more idiomatic.

After day 3, read other people's Go: the standard library, and a mid-sized project such as `caddy`, `hugo` or `gitea`. That is where idioms really sink in.
