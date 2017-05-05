#!/bin/bash

# The root of the build/dist directory
readonly OSCAN_ROOT=$(
  unset CDPATH
  root=$(dirname "${BASH_SOURCE}")/..

  cd "${root}"
  root=`pwd`
  if [ -h "${root}" ]; then
    readlink "${root}"
  else
    pwd
  fi
)

readonly OSCAN_GOPATH=$(
  unset CDPATH
  cd ${OSCAN_ROOT}/../../../..
  pwd
)
source "${CODE_ROOT}/hack/lib/util/environment.sh"

readonly OSCAN_GO_PACKAGE=github.com/openshift/$PROJECT
readonly OSCAN_OUTPUT_SUBPATH="${OSCAN_OUTPUT_SUBPATH:-_output/local}"
readonly OSCAN_OUTPUT="${OSCAN_ROOT}/${OSCAN_OUTPUT_SUBPATH}"
readonly OSCAN_OUTPUT_BINPATH="${OSCAN_OUTPUT}/bin"

# oscan::build::setup_env will check that the `go` commands is available in
# ${PATH}. If not running on Travis, it will also check that the Go version is
# good enough for the webdav code requirements (1.5+).
#
# Output Vars:
#   export GOPATH - A modified GOPATH to our created tree along with extra
#     stuff.
#   export GOBIN - This is actively unset if already set as we want binaries
#     placed in a predictable place.
function oscan::build::setup_env() {
  if [[ -z "$(which go)" ]]; then
    cat <<EOF

Can't find 'go' in PATH, please fix and retry.
See http://golang.org/doc/install for installation instructions.

EOF
    exit 2
  fi

  # Travis continuous build uses a head go release that doesn't report
  # a version number, so we skip this check on Travis.  It's unnecessary
  # there anyway.
  if [[ "${TRAVIS:-}" != "true" ]]; then
    local go_version
    go_version=($(go version))
    if [[ "${go_version[2]}" < "go1.5" ]]; then
      cat <<EOF

Detected Go version: ${go_version[*]}.
$PROJECT builds require Go version 1.5 or greater.

EOF
      exit 2
    fi
  fi

  unset GOBIN

  export GOPATH=${OSCAN_ROOT}/Godeps/_workspace:${OSCAN_GOPATH}
  export OSCAN_TARGET_BIN=${OSCAN_GOPATH}/bin
}

# oscan::build::get_version_vars loads the standard version variables as
# ENV vars
function oscan::build::get_version_vars() {
  local git=(git --work-tree "${OSCAN_ROOT}")

  OSCAN_GIT_COMMIT="${OSCAN_GIT_COMMIT-}"
  OSCAN_GIT_TREE_STATE="${OSCAN_GIT_TREE_STATE-}"
  OSCAN_GIT_VERSION="${OSCAN_GIT_VERSION-}"
  OSCAN_GIT_MAJOR="${OSCAN_GIT_MAJOR-}"
  OSCAN_GIT_MINOR="${OSCAN_GIT_MINOR-}"

  if [[ -n ${OSCAN_GIT_COMMIT-} ]] || OSCAN_GIT_COMMIT=$("${git[@]}" rev-parse --short "HEAD^{commit}" 2>/dev/null); then
    if [[ -z ${OSCAN_GIT_TREE_STATE-} ]]; then
      # Check if the tree is dirty.  default to dirty
      if git_status=$("${git[@]}" status --porcelain 2>/dev/null) && [[ -z ${git_status} ]]; then
        OSCAN_GIT_TREE_STATE="clean"
      else
        OSCAN_GIT_TREE_STATE="dirty"
      fi
    fi
    # Use git describe to find the version based on annotated tags.
    if [[ -n ${OSCAN_GIT_VERSION-} ]] || OSCAN_GIT_VERSION=$("${git[@]}" describe --long --tags --abbrev=7 "${OSCAN_GIT_COMMIT}^{commit}" 2>/dev/null); then
      # Try to match the "git describe" output to a regex to try to extract
      # the "major" and "minor" versions and whether this is the exact tagged
      # version or whether the tree is between two tagged versions.
      if [[ "${OSCAN_GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)*([-].*)?$ ]]; then
        OSCAN_GIT_MAJOR=${BASH_REMATCH[1]}
        OSCAN_GIT_MINOR=${BASH_REMATCH[2]}
        if [[ -n "${BASH_REMATCH[4]}" ]]; then
          OSCAN_GIT_MINOR+="+"
        fi
      fi

      # This translates the "git describe" to an actual semver.org
      # compatible semantic version that looks something like this:
      #   v1.1.0-alpha.0.6+84c76d1-345
      OSCAN_GIT_VERSION=$(echo "${OSCAN_GIT_VERSION}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{7,40\}\)$/\+\2-\1/")
      # If this is an exact tag, remove the last segment.
      OSCAN_GIT_VERSION=$(echo "${OSCAN_GIT_VERSION}" | sed "s/-0$//")
      if [[ "${OSCAN_GIT_TREE_STATE}" == "dirty" ]]; then
        # git describe --dirty only considers changes to existing files, but
        # that is problematic since new untracked .go files affect the build,
        # so use our idea of "dirty" from git status instead.
        OSCAN_GIT_VERSION+="-dirty"
      fi
    fi
  fi
}
readonly -f oscan::build::get_version_vars

# golang 1.5 wants `-X key=val`, but golang 1.4- REQUIRES `-X key val`
function oscan::build::ldflag() {
  local key=${1}
  local val=${2}

  GO_VERSION=($(go version))
  if [[ -n $(echo "${GO_VERSION[2]}" | grep -E 'go1.4') ]]; then
    echo "-X ${key} ${val}"
  else
    echo "-X ${key}=${val}"
  fi
}
readonly -f oscan::build::ldflag

# oscan::build::ldflags calculates the -ldflags argument for building OpenShift
function oscan::build::ldflags() {
  # Run this in a subshell to prevent settings/variables from leaking.
  set -o errexit
  set -o nounset
  set -o pipefail

  cd "${OSCAN_ROOT}"

  oscan::build::get_version_vars

  local buildDate="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

  declare -a ldflags=()

  ldflags+=($(oscan::build::ldflag "${OSCAN_GO_PACKAGE}/pkg/version.majorFromGit" "${OSCAN_GIT_MAJOR}"))
  ldflags+=($(oscan::build::ldflag "${OSCAN_GO_PACKAGE}/pkg/version.minorFromGit" "${OSCAN_GIT_MINOR}"))
  ldflags+=($(oscan::build::ldflag "${OSCAN_GO_PACKAGE}/pkg/version.versionFromGit" "${OSCAN_GIT_VERSION}"))
  ldflags+=($(oscan::build::ldflag "${OSCAN_GO_PACKAGE}/pkg/version.commitFromGit" "${OSCAN_GIT_COMMIT}"))
  ldflags+=($(oscan::build::ldflag "${OSCAN_GO_PACKAGE}/pkg/version.buildDate" "${buildDate}"))

  # The -ldflags parameter takes a single string, so join the output.
  echo "${ldflags[*]-}"
}
readonly -f oscan::build::ldflags

# Build binary.
function oscan::build::build_binaries() {
  # Create a sub-shell so that we don't pollute the outer environment
  (
    # Check for `go` binary and set ${GOPATH}.
    oscan::build::setup_env

    # Fetch the version.
    local version_ldflags
    version_ldflags=$(oscan::build::ldflags)

    # Making this super simple for now.
    local platform="local"
    export GOBIN="${OSCAN_OUTPUT_BINPATH}/${platform}"

    mkdir -p "${OSCAN_OUTPUT_BINPATH}/${platform}"
    for cmd in cmd/*; do
      [ ! -d "${cmd}" ] || ( cd "${cmd}"; go install -ldflags="${version_ldflags}")
    done
  )
}
readonly -f oscan::build::build_binaries
