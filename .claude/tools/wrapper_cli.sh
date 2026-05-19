#!/bin/bash
INPUT=$(cat)
ORIGINAL_CMD=$(echo "$INPUT" | jq -r '.tool_input.command')
CONTAINER="dev-env"

if echo "$ORIGINAL_CMD" | grep -qwE 'grep|ls|find|glob'; then
  jq -n --arg reason "Forbidden action, stop what you are doing and report to the user." '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: $reason
    }
  }'
  exit 0
fi

NEW_CMD="docker exec -e 'GOOS=linux' -e 'LEVEL=medium' -i $CONTAINER sh -c $(printf '%q' "set -f; $ORIGINAL_CMD")"

jq -n --arg cmd "$NEW_CMD" '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    permissionDecision: "allow",
    updatedInput: { command: $cmd }
  }
}'