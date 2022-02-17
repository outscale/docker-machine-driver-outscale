#!/bin/bash
## From https://raw.githubusercontent.com/docker/machine/master/test/integration/helpers.bash

function echo_to_log {
    echo "$BATS_TEST_NAME
----------
$output
----------

"   >> ${BATS_LOG}
}

function teardown {
    echo_to_log
}

function errecho {
    >&2 echo "$@"
}

function only_if_env {
    if [[ ${!1} != "$2" ]]; then
        errecho "This test requires the $1 environment variable to be set to $2. Skipping..."
        skip
    fi
}

function require_env {
    if [[ -z ${!1} ]]; then
        errecho "This test requires the $1 environment variable to be set in order to run."
        exit 1
    fi
}

function require_tool {
    set +e 
    which $1
    if [ $? -ne 0 ]; then
		errecho "This test requires $1 to be installed"
		exit 1
	fi
    set -e
}

function use_disposable_machine {
    if [[ -z "$NAME" ]]; then
        export NAME="bats-outscale-test-$(date +%s)"
    fi
}

function use_shared_machine {
    if [[ -z "$NAME" ]]; then
      export NAME="$SHARED_NAME"
      if [[ $(machine ls -q --filter name=$NAME | wc -l) -eq 0 ]]; then
          machine create -d outscale $NAME &>/dev/null
      fi
    fi
}

# Make sure these aren't set while tests run (can cause confusing behavior)
unset DOCKER_HOST DOCKER_TLS_VERIFY DOCKER_CERT_DIR