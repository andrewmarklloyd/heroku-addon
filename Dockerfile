FROM alpine

COPY bin/heroku-addon /app/
WORKDIR /app

ENTRYPOINT ["/app/heroku-addon"]
