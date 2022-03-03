#!/usr/bin/env bats

function create_osc_cli_default_profile {
	require_env OSC_ACCESS_KEY
	require_env OSC_SECRET_KEY
	require_env OSC_REGION

	require_tool osc-cli
	mkdir -p $HOME/.osc
	cat <<EOF > $HOME/.osc/config.json
	{ 
		"default":
		{
			"access_key": "$OSC_ACCESS_KEY",
			"secret_key": "$OSC_SECRET_KEY",
			"host": "outscale.com",
			"https": true,
			"method": "POST",
			"region_name": "$OSC_REGION"
		}
	}
EOF
}

function init_resources {
    require_env OSC_ACCESS_KEY
	require_env OSC_SECRET_KEY
	require_env OSC_REGION

	require_tool terraform

    export TF_VAR_access_key_id=$OSC_ACCESS_KEY
    export TF_VAR_secret_key_id=$OSC_SECRET_KEY
    export TF_VAR_region=$OSC_REGION

    terraform -chdir=test/resources/security_groups init 
    terraform -chdir=test/resources/security_groups apply --auto-approve 
    export SG_ID=$(terraform -chdir=test/resources/security_groups show -json | jq -r '.values.root_module.resources[] | select(.address=="outscale_security_group.test-docker") | .values.id')
}

function destroy_resources {
    require_env OSC_ACCESS_KEY
	require_env OSC_SECRET_KEY
	require_env OSC_REGION

	require_tool terraform

    export TF_VAR_access_key_id=$OSC_ACCESS_KEY
    export TF_VAR_secret_key_id=$OSC_SECRET_KEY
    export TF_VAR_region=$OSC_REGION

    terraform -chdir=test/resources/security_groups init 
    terraform -chdir=test/resources/security_groups destroy --auto-approve 
}


setup_file() {
	load ${BASE_TEST_DIR}/helpers.bash

	use_disposable_machine

	init_resources
    create_osc_cli_default_profile

	require_env OSC_ACCESS_KEY
	require_env OSC_SECRET_KEY
}

teardown() {
    load ${BASE_TEST_DIR}/helpers.bash

    machine rm --force $NAME
}

teardown_file() {
	load ${BASE_TEST_DIR}/helpers.bash

	destroy_resources
}


@test "Test with incorect security group id" {
	run machine create -d outscale --outscale-security-group-ids "toto" $NAME
	[ "$status" -eq 1 ]
}

@test "Tests with good SG" {
	run machine create -d outscale --outscale-security-group-ids $SG_ID $NAME
	[ "$status" -eq 0 ]

    export VMID=$(machine inspect $NAME | jq -r '.Driver.VmId')

	res=$(osc-cli api ReadVms --Filters "{\"VmIds\": [\"$VMID\"]}" | jq -r --arg vm_id $VMID --arg sg_id $SG_ID '.Vms[] | select(.VmId==$vm_id) | .SecurityGroups[] | select(.SecurityGroupId==$sg_id)')
	[ "$res" != "" ]
}