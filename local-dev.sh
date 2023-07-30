#!/bin/bash

set -e

export PORT=8080

if [[ ${1} != 'skip-front' ]]; then
    cd frontend
    npm run build
    cd ../
fi


op run --env-file=".env.server-staging.tmpl" -- go run *.go
