---
name: test-request
description: Test a feature
disable-model-invocation: true
#context: fork
#agent: Explore
---

Use the skill test-request-management

Herei is the list of available tests:
BEGIN
!`ls -al .agent/requests/test/*.md`
END

Test the feature, the technical name is "$ARGUMENTS".
If the technical name is known, just load the task.
Else ask the user for it plus ask him what feature he wants test.

