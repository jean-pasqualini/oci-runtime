---
name: feature-request-list
description: List feature requests
disable-model-invocation: true
effort: low
#context: fork
#agent: Explore
---

Say "EFFORT LEVEL: ${CLAUDE_EFFORT}".

Here is the list of feature requests
> BEGIN
!`cat .agent/requests/feature/*.md`
> END
Show as a table with technical-name, description, how much task ouf of total task are completed.