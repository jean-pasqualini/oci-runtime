#!/bin/bash
INPUT=$(cat)
ORIGINAL_CMD=$(echo "$INPUT" | jq -r '.tool_input.command')
CONTAINER="dev-env"

NEW_CMD="docker exec -e 'GOOS=linux' -e 'LEVEL=medium' -i $CONTAINER sh -c $(printf '%q' "$ORIGINAL_CMD")"

jq -n --arg cmd "$NEW_CMD" '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    permissionDecision: "allow",
    updatedInput: { command: $cmd }
  }
}'