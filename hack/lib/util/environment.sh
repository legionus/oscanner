#!/bin/bash

# see https://github.com/openshift/origin/blob/master/hack/lib/util/environment.sh

# This script holds library functions for setting up the shell environment for OpenShift scripts
#
# This script assumes $OSCAN_ROOT is set before being sourced
source "${OSCAN_ROOT}/hack/util.sh"

export PROJECT=oscanner

# oscan::util::environment::setup_time_vars sets up environment variables that describe durations of time
# These variables can be used to specify times for other utility functions
#
# Globals:
#  None
# Arguments:
#  None
# Returns:
#  - export TIME_MS
#  - export TIME_SEC
#  - export TIME_MIN
function oscan::util::environment::setup_time_vars() {
    export TIME_MS=1
    export TIME_SEC="$(( 1000  * ${TIME_MS} ))"
    export TIME_MIN="$(( 60 * ${TIME_SEC} ))"
}

# oscan::util::environment::setup_all_server_vars sets up all environment variables necessary to configure and start an OpenShift server
#
# Globals:
#  - OSCAN_ROOT
#  - PATH
#  - TMPDIR
#  - LOG_DIR
#  - ARTIFACT_DIR
# Arguments:
#  - 1: the path under the root temporary directory for OpenShift where these subdirectories should be made
# Returns:
#  - export PATH
#  - export BASETMPDIR
#  - export LOG_DIR
#  - export VOLUME_DIR
#  - export ARTIFACT_DIR
function oscan::util::environment::setup_all_server_vars() {
    local subtempdir=$1

    oscan::util::environment::update_path_var
    oscan::util::environment::setup_tmpdir_vars "${subtempdir}"
}

# oscan::util::environment::update_path_var updates $PATH so that OpenShift binaries are available
#
# Globals:
#  - OSCAN_ROOT
#  - PATH
# Arguments:
#  None
# Returns:
#  - export PATH
function oscan::util::environment::update_path_var() {
    export PATH="${OSCAN_ROOT}/_output/local/bin/$(oscan::util::host_platform):${PATH}"
}

# oscan::util::environment::setup_misc_tmpdir_vars sets up temporary directory path variables
#
# Globals:
#  - TMPDIR
#  - LOG_DIR
#  - ARTIFACT_DIR
# Arguments:
#  - 1: the path under the root temporary directory for OpenShift where these subdirectories should be made
# Returns:
#  - export BASETMPDIR
#  - export LOG_DIR
#  - export VOLUME_DIR
#  - export ARTIFACT_DIR
#  - export FAKE_HOME_DIR
#  - export HOME
function oscan::util::environment::setup_tmpdir_vars() {
    local sub_dir=$1

    export BASETMPDIR="${TPMDIR:-/tmp}/$PROJECT/${sub_dir}"
    export LOG_DIR="${LOG_DIR:-${BASETMPDIR}/logs}"
    export VOLUME_DIR="${BASETMPDIR}/volumes"
    export ARTIFACT_DIR="${ARTIFACT_DIR:-${BASETMPDIR}/artifacts}"

    # change the location of $HOME so no one does anything naughty
    export FAKE_HOME_DIR="${BASETMPDIR}/openshift.local.home"
    export HOME="${FAKE_HOME_DIR}"

    mkdir -p  "${BASETMPDIR}" "${LOG_DIR}" "${VOLUME_DIR}" "${ARTIFACT_DIR}" "${HOME}"
}
