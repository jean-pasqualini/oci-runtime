---
name: feature-request-management
description: Feature Request Management
disable-model-invocation: false
user-invocable: false
#context: fork
#agent: Explore
---

This describes how to handle a feature request from the user.
Please ask a technical name for that feature request that you can refer to follow the task across sessions.
Refuse sharply any feature request that doesn't makes sense in that project.

If that's a new task:

1. Discovery pass — BEFORE splitting into tasks:
   - Read the relevant code paths to understand existing conventions, port shapes, and analogous features already in the codebase.
   - List the architectural / convention decisions the work hinges on — things knowable upfront, not after coding starts.
   - Ask the user those decisions in a single batched `AskUserQuestion` call. Only ask what genuinely needs user judgement — skip anything derivable from the code.
   - Record answers verbatim in a `decisions:` section of the markdown file before writing tasks.
2. Split the work into tasks that respect the recorded decisions. Each task should be small enough to land in one pass without surfacing new architectural questions.
3. If a task DOES surface a decision the discovery pass missed, pause, ask the user, append to `decisions:`, then continue.
4. Describe the result in `.agent/requests/feature/<technical-name>.md` using the template at `.claude/skills/feature-request-management/templates/feature-request.md`.
5. Use the checkbox format for task items.
6. After a task is completed, update the task item and add a line of memory.

If that's an existing task,
Please read the related markdown task file