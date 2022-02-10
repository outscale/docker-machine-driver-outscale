#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

use_disposable_machine

require_env OSC_ACCESS_KEY
require_env OSC_SECRET_KEY

@test "Creation with wrong instance type" {
    run machine create -d outscale --outscale-instance-type toto $NAME 
    echo ${output}
    [ "$status" -eq 1 ]
}