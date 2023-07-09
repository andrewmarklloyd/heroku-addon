FROM alpine

COPY bin/heroku-addon /app/
COPY frontend/build /app/frontend/build

WORKDIR /app

ENTRYPOINT ["/app/heroku-addon"]
