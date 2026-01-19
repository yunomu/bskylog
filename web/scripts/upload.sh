#!/bin/sh

set -eu

if [ $# -ne 1 ]; then
	echo "Usage: $0 <STACK_NAME>"
	exit 1
fi

STACK_NAME=$1

BUCKET_NAME=$(aws cloudformation describe-stacks --stack-name "${STACK_NAME}" --query "Stacks[0].Outputs[?OutputKey=='PublishBucket'].OutputValue" --output text)

if [ -z "${BUCKET_NAME}" ]; then
	echo "Error: PublishBucket not found in stack: ${STACK_NAME}" >&2
	exit 1
fi

echo "Uploading to bucket: ${BUCKET_NAME}"

aws s3 sync ./public "s3://${BUCKET_NAME}/"

echo "Upload completed successfully."
