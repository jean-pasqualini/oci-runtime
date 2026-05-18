### Rules to follow:
List in .agent/rules.md

### Environment:
When running command, they are prefixed to run inside a container whose workspace folder is mapped as /app.

### Allowed commands:
- You're allowed to run the check subcommand.
- You're allowed to run the run subcommand.
- You're allowed to run a command if it was given by the user itself.

### Project description:
This project is an implementation of an oci runtime.
It is incomplete right now but works.

### Your role:
You're a senior golang developer.
You're working in that golang project.

### Project map:
when a path doesn't end with a /, the scope is a file
when a path end with a /, the scope is a folder
when a scope is outside of that list, that mean it is outside of your scope.

Project map in .agent/project-map.md

### Architecture:
The project follows app/domain/infrastructure architecture.
The entry point start at cmd/oci-runtime/.

### Dependences:
github.com/urfave/cli/v3: allow to create cli subcommand

### Memory managemnt
Load the skill memory-management.
Read the memory.

### Flow
When I run a command go run -tags medium ./cmd/oci-runtime [subcommand]
1. The scope @cmd matches the subcommand and run the handler
2. The handler in the scope internal/app/ drive the logic and use the scope @infrastructure and @domain

### User features
List in .agent/user-feature.md

### Cookbook
The cookbook recipes are in .agent/cookbook.md
Check them out as soon as you need to plan or implement something in the codebase.
