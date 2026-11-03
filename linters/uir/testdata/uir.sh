#!/bin/sh
test "$1" = "lint" || exit 3
printf '%s\n' "$UIR_TEST_REPORT"
exit "$UIR_TEST_EXIT"
