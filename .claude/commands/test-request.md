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

Herei is the list of feature the use could write a test for:
BEGIN
!`ls -al .agent/requests/feature/*.md`
END

Test the feature, the technical name is "$ARGUMENTS".
If the technical name is known, just load the task.
Else ask the user for what feature-request he wants to test.

