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


setup_file() {
	load ${BASE_TEST_DIR}/helpers.bash

	use_disposable_machine

	create_osc_cli_default_profile

	require_env OSC_ACCESS_KEY
	require_env OSC_SECRET_KEY
}

@test "Test all extra tags" {
	run machine create -d outscale --outscale-extra-tags-all "all1=valueall1" --outscale-extra-tags-all "all2=valueall2" --outscale-extra-tags-instances "vm1=valuevm1" --outscale-extra-tags-all "vm2=valuevm2" $NAME
	echo ${output}
	[ "$status" -eq 0 ]

	export VMID=$(machine inspect $NAME | jq -r '.Driver.VmId')
	export SGID=$(machine inspect $NAME | jq -r '.Driver.SecurityGroupId')

	res=$(osc-cli api ReadVms --Filters "{ \"VmIds\" : [\"$VMID\"], \"Tags\": [\"all1=valueall1\", \"all2=valueall2\", \"vm1=valuevm1\", \"vm2=valuevm2\"] }" | jq '.Vms | length')
	[ "$res" == 1 ]

	res=$(osc-cli api ReadSecurityGroups --Filters "{ \"SecurityGroupIds\" : [\"$SGID\"], \"Tags\": [\"all1=valueall1\", \"all2=valueall2\"] }" | jq '.SecurityGroups | length')
	[ "$res" == 1 ]

	run machine rm -f $NAME
	[ "$status" -eq 0 ]
}

@test "Tests wrong format of tags" {
	run machine create -d outscale --outscale-extra-tags-all "all1valueall1" $NAME
	[ "$status" -eq 1 ]

	run machine create -d outscale --outscale-extra-tags-all "=all1valueall1" $NAME
	[ "$status" -eq 1 ]

	run machine create -d outscale --outscale-extra-tags-all "all1=value=all1" $NAME
	[ "$status" -eq 1 ]
}