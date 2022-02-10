#!/usr/bin/env bats

setup_file() {
    load ${BASE_TEST_DIR}/helpers.bash

    use_disposable_machine

    require_env OSC_ACCESS_KEY
    require_env OSC_SECRET_KEY
}

@test "Default creation" {
    run machine create -d outscale $NAME
    echo ${output}
    [ "$status" -eq 0 ]
}

@test "Default stop" {
    run machine stop $NAME
    echo ${output}
    [ "$status" -eq 0 ]
}

@test "Default start" {
    run machine start $NAME
    echo ${output}
    [ "$status" -eq 0 ]
}

@test "Default remove" {
    run machine rm -y $NAME
    echo ${output}
    [ "$status" -eq 0 ]
}