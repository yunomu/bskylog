#!/bin/sh

set -eu

if [ $# -ne 1 ]; then
	echo "Usage: $0 <STACK_NAME>"
	exit 1
fi

STACK_NAME=$1

DISTRIBUTION_ID=$(aws cloudformation describe-stacks --stack-name "${STACK_NAME}" --query "Stacks[0].Outputs[?OutputKey=='Distribution'].OutputValue" --output text)

if [ -z "${DISTRIBUTION_ID}" ]; then
	echo "Error: Distribution not found in stack: ${STACK_NAME}" >&2
	exit 1
fi

echo "Invalidating cache for distribution: ${DISTRIBUTION_ID}"

aws cloudfront create-invalidation --distribution-id "${DISTRIBUTION_ID}" --paths "/index.html" "/main.js" "/config.json"

echo "Invalidation request sent successfully."
