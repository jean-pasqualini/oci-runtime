### Project map

go.mod -> the list of dependency
cmd/oci-runtime/ -> the entrypoint
internal/ -> the codebase
internal/app/ -> the handlers
internal/app/mw/ -> the middlewares
internal/domain/ -> some shared domain model
infrastructure/linux/ -> every thing tight to linux (mount, network, ns, proc)
infrastructure/technical/ -> technical reusable accross projects (config, logging, error)
infrastructure/transport/ipc/ -> communication inter process (ipc)
.claude/ -> ai agent related (memory, commands, rules, tasks, templates, tools, ...)