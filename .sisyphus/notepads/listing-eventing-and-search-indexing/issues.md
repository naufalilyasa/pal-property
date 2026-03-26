# Issues

- Full `go test ./...` is currently blocked in this environment because Docker/testcontainers cannot connect to `/var/run/docker.sock`. Targeted eventing compile/test slice (`./internal/service`, `./cmd/property-service`, `./pkg/...`) is green, but handler/migrate suites that require testcontainers must be rerun in a Docker-enabled environment before final sign-off.
- Final verification remains environment-blocked for the same reason: `cmd/migrate` and `internal/handler/http` suites cannot complete here without Docker, so the eventing/search plan is code-ready but not fully sign-off-ready under the current runner.
