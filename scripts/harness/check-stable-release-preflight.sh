#!/bin/sh

set -eu
script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)

# No arguments runs offline tests. Live modes read forge state or local assets;
# only the separately authorized workflow may create a tag or release.
exec python3 "$script_dir/check-stable-release-preflight.py" "$project_root" "$@"
