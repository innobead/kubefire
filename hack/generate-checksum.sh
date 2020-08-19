#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
#set -o xtrace

find $1 -print0 | while IFS= read -r -d '' f; do
    if [[ -f "$f" ]] && [[ -x "$f" ]] ; then
       shasum -a 256 "$f" > "$f.sha256";
    fi
done