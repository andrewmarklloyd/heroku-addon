#!/bin/bash

export PORT=8080

cd frontend
npm run build
cd ../

op run --env-file=".env.server-staging.tmpl" -- go run *.go
