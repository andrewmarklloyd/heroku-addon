#!/bin/bash


export HEROKU_ADDON_PASSWORD=$(cat addon-manifest.json | jq -r '.api.password')
export HEROKU_ADDON_USERNAME=$(cat addon-manifest.json | jq -r '.id')
export ENCRYPTION_KEY=$(op read op://heroku-addon/config/ENCRYPTION_KEY)
export HEROKU_CLIENT_SECRET=$(op read op://heroku-addon/config/HEROKU_CLIENT_SECRET)
export PORT=8080

go run *.go