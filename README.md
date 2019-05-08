# `gracego` enables gracefully restart or upgrade golang application.

## Usage

### Start server
Your server must implement the interface `GraceServer` with two methods:
```go
type GraceServer interface {
	Serve(listener net.Listener) error
	Shutdown(ctx context.Context) error
}
```

Use `gracego.Serve()` to start your server:
```go
func main() {
	server = &http.Server{}

	err := gracego.Serve(server, "demo", ":8080")
	if err != nil {
		fmt.Printf("server error: %v\n", err)
	}
}
```

### Restart server gracefully

```bash
kill -HUP <PID>
```

### Upgrade server gracefully

- `v2`: the new version to upgrade
- `echo`: the relative path of the upgrade command in the download.zip
- `http://127.0.0.1:8081/download.zip`: the upgrade url, which must be a zip file and end with `.zip` or `.jar`.
```go

err := gracego.Upgrade("v2", "echo", "http://127.0.0.1:8081/download.zip")
if err != nil {
    // error handle
}
```

## Examples

- [echo](examples/echo/README.md): example to shutdown, restart and upgrade gracefully
