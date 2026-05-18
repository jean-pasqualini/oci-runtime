### Cookbook

#### How to use the logger
```go
l := logging.FromContext(ctx)
l.Info("something")
l.Error("something")
l.Debug("something")
l.With("key", "value").Info("something")
```

### How to wrap error with contextual information
There is a package "oci-runtime/internal/infrastructure/technical/xerr"
It expose a function Op() that wrap an error with contextual information
Usage:
```go
xerr.Op("operation you were doing", err, xerr.KV{
			"key": value,
})
```

#### How to add a subcommand ? 
in the scope @cmd/oci-runtime/,
you need to modify the file main.go and cmd.go, you're allowed to read/edit them.

#### How to add a handler ? 
Look at the scope @internal/app/example.go as an example.

#### What can the handler uses ?
The handler can use every interfaces exposes by @internal/app/ports.go
If something is missing to implement a handler, you can add it to ports.go then in another task implement it in the infrastructure part.
Those interfaces, allow you to know what methods are available in infrastructure without having to explore yourself.