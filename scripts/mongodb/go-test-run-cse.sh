#!/bin/bash

# To see the profile names: cat ~/.aws/config
# For more information: https://github.com/mongodb-labs/drivers-evergreen-tools/blob/master/.evergreen/secrets_handling/README.md

# Check if DRIVER_PATH is set
if [ -z "$DRIVER_PATH" ]; then
  echo "Error: DRIVER_PATH environment variable is not set."
  exit 1
fi

aws sso login --profile $AWS_PROFILE
#

bash $DRIVERS_TOOLS/.evergreen/csfle/setup-secrets.sh
source secrets-export.sh

#cd $DRIVER_PATH || exit
#
#go test ./internal/integration -run=$1 -v -tags=cse -failfast
##go test ./internal/integration/unified -run=$1 -v -tags=cse -failfast
#
#rm -rf secrets-export.sh

