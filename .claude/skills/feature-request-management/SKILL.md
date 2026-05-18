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

If that's a new task,
Please split this in multiple task, and describe this in the markdown file .agent/requests/feature/
Use the following template .claude/templates/feautre-request.md
Use the checkbox format for task item.
After a task is completed, update the task item and add a line of memory.

If that's an existing task,
Please read the related markdown task file