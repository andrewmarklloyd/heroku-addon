#!/bin/bash


export HEROKU_ADDON_PASSWORD=$(cat addon-manifest.json | jq -r '.api.password')
export HEROKU_ADDON_USERNAME=$(cat addon-manifest.json | jq -r '.id')

eval $(echo "export ENCRYPTION_KEY=op://heroku-addon/config/ENCRYPTION_KEY
export DATABASE_URL=op://heroku-addon/config/DATABASE_URL
export SESSION_SECRET_HASH_KEY=op://heroku-addon/config/SESSION_SECRET_HASH_KEY
export SESSION_SECRET_ENCRYPTION_KEY=op://heroku-addon/config/SESSION_SECRET_ENCRYPTION_KEY
export HEROKU_CLIENT_SECRET=op://heroku-addon/config/HEROKU_CLIENT_SECRET
export GITHUB_CLIENT_ID=op://heroku-addon/config/GITHUB_CLIENT_ID
export GITHUB_CLIENT_SECRET=op://heroku-addon/config/GITHUB_CLIENT_SECRET
export GITHUB_REDIRECT_URI=op://heroku-addon/config/GITHUB_REDIRECT_URI" | op inject)

export PORT=8080

cd frontend
npm run build
cd ../

go run *.go