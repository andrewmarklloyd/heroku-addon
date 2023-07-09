#!/bin/bash


export ADDON_PASSWORD=$(cat addon-manifest.json | jq -r '.api.password')
export ADDON_USERNAME=$(cat addon-manifest.json | jq -r '.id')
export ENCRYPTION_KEY=$(op read op://heroku-addon/config/ENCRYPTION_KEY)
export CLIENT_SECRET=$(op read op://heroku-addon/config/CLIENT_SECRET)
export ADDON_NAME=alloyd-poc

go run *.go