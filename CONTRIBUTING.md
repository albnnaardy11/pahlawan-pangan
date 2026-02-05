# Contributing to Pahlawan Pangan

So you want to contribute. Good. But read this first, because I have very 
low tolerance for bad code and even lower tolerance for people who don't 
follow basic standards.

## The Golden Rule

**Don't send me crap.**

If your code is messy, undocumented, or lacks tests, I will ignore your pull 
request. My time is better spent fixing real issues than cleaning up after 
you.

## Coding Standards

We use Go. Go has a standard. Follow it.

1.  **`go fmt` is not optional.** If your code isn't formatted correctly, the
    CI will kill it. I won't even look at it.
2.  **`go vet` and `golangci-lint` must pass.** If you're ignoring lint 
    errors, you're doing it wrong.
3.  **No Goroutine Leaks.** If you start a goroutine, ensure it has a way 
    to exit. Use `context.Context` properly. I will check for this.
4.  **Error Handling.** Don't ignore errors. Don't use `panic` unless the 
    entire world is ending. Handle errors gracefully.

## Performance is a Feature

Pahlawan Pangan is built for high-scale (10M+ TPS). If you add a feature 
that slows down the matching engine, I don't want it. 

- Use `sync.Pool` for high-frequency allocations.
- Avoid unnecessary JSON marshaling in hot paths.
- If you're unsure about performance, run a benchmark (`go test -bench`). 
  Include the benchmark results in your PR.

## Testing

If you change the matching logic, you must add tests. If you don't, I will
assume your code is broken.

- **Unit Tests**: Mandatory for all logic in `internal/`.
- **Race Detector**: Always run your tests with `-race`. If it fails, fix the 
  concurrency bug before even thinking about pushing.

## Pull Request Process

1.  **Atomic Commits**: Each commit should do one thing and do it well. 
    Don't send me a "Catch-all" commit with 5,000 lines of changes.
2.  **Descriptive Messages**: Tell me *why* you're doing something, not just 
    *what*. The code tells me what, the message tells me the reasoning.
3.  **Rebase, don't Merge**: Keep the git history clean. Linear history is 
    far superior to a spaghetti mess of merge commits.

## Final Note

I care about code that works, code that scales, and code that saves lives. 
If you can provide that, you're welcome here. If you just want to add 
"cool" frameworks or bloat, go somewhere else.

---
*Keep it clean. Keep it fast. Or don't send it at all.*
