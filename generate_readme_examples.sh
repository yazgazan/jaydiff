#!/usr/bin/env bash

cat << EOF
Getting a full diff of two json files:

$(echo '```sh')
$ jaydiff --show-types old.json new.json

$(./jaydiff --indent='    ' --show-types test_files/lhs.json test_files/rhs.json)
$(echo '```')

Ignoring fields:

$(echo '```sh')
$ jaydiff --show-types \\
	  --ignore='.b\[\]' --ignore='.d' --ignore='.c.[ac]' \\
	    old.json new.json

$(./jaydiff --indent='    ' --show-types \
	--ignore='.b\[\]' --ignore='.d' --ignore='.c.[ac]' \
	test_files/lhs.json test_files/rhs.json
)
$(echo '```')

Report format:

$(echo '```sh')
$ jaydiff --report --show-types old.json new.json

$(./jaydiff --report --indent='    ' --show-types test_files/lhs.json test_files/rhs.json)
$(echo '```')
EOF
