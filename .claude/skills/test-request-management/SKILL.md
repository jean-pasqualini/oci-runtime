---
name: test-request-management
description: Test Request Management
disable-model-invocation: false
user-invocable: false
#context: fork
#agent: Explore
---

This describes how to handle a run a test.
Follow the test plan in .agent/test/[feature-request-name]
Don't check the code, you're only focused on testing.
Ask the user if he wants to only executed the non-passed/non-executed onces.
Add a line or update the existant one to the Execution log of the test plan after a test case tested has been tested before going for the next one.
At first fail test, stop the testing and report to the user.