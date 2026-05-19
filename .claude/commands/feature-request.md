---
name: feature-request
description: Request a feature
disable-model-invocation: true
#context: fork
#agent: Explore
---

Use the skill feature-request-management.

Herei is the list of available features:
BEGIN
!`ls -al .agent/requests/feature/*.md`
END

Request the feature, the technical name is "$ARGUMENTS".
If the technical name is known, just load the task.
Else ask the user for it plus ask him what feature he wants.
Show the list of task to do as checkbox.

