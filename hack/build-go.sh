#!/bin/bash

# This script sets up a go workspace locally and builds all go components.

set -o errexit
set -o nounset
set -o pipefail

STARTTIME=$(date +%s)
CODE_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${CODE_ROOT}/hack/util.sh"
source "${CODE_ROOT}/hack/common.sh"
oscan::log::install_errexit

oscan::build::build_binaries "$@"

ret=$?; ENDTIME=$(date +%s); echo "$0 took $(($ENDTIME - $STARTTIME)) seconds"; exit "$ret"
