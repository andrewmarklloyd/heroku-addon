#!/bin/bash

set -e

export PORT=8080
export CI=true
export REACT_APP_STRIPE_PUBLIC_KEY=$(op read op://heroku-addon/config/REACT_APP_STRIPE_PUBLIC_KEY)
export TEST_MODE=true

if [[ ${1} != 'skip-front' ]]; then
    cd frontend
    npm run build
    cd ../
fi


op run --env-file=".env.server-staging.tmpl" -- go run *.go
