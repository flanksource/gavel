#!/bin/sh
set -eu
test "$1" = lint
test "$2" = --format
for fragment in .uir.d/gavel-*.yaml
do
    cp "$fragment" "$UIR_TEST_CAPTURE"
done
printf '%s\n' '[{"linter":"uir","success":true,"violations":[]}]'
