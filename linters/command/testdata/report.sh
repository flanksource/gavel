#!/bin/sh
printf '%s\n' "$*" > "$COMMAND_TEST_ARGS"
printf '%s\n' "$COMMAND_TEST_REPORT"
exit "$COMMAND_TEST_EXIT"
