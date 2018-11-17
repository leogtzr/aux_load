#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

readonly working_dir=$(dirname $(readlink --canonicalize-existing "${0}"))
cd "${working_dir}"

go clean
find "${working_dir}" -type f -name 'aux_load.log*' -print -delete

exit 0