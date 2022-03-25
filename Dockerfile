FROM docker.io/library/alpine:3.14 as os

# install ca-certificates
RUN apk add --update --no-cache ca-certificates

# create www-data
RUN set -x ; \
  addgroup -g 82 -S www-data ; \
  adduser -u 82 -D -S -G www-data www-data && exit 0 ; exit 1

# build the frontend
FROM docker.io/library/node:16-bullseye-slim as frontend
RUN apt-get update && apt-get -y install build-essential python3
ADD cmd/odbcd/ /app/cmd/odbcd/
WORKDIR /app/cmd/odbcd/
RUN yarn install --frozen-lockfile
RUN yarn dist

# build the backend
FROM docker.io/library/golang:1.18 as builder
ADD . /app/
WORKDIR /app/
COPY --from=frontend /app/cmd/odbcd/dist /app/cmd/odbcd/dist
RUN go get ./cmd/odbcd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o odbcd ./cmd/odbcd

# add it into a scratch image
FROM docker.io/library/alpine:3.14

# add the user
COPY --from=os /etc/passwd /etc/passwd
COPY --from=os /etc/group /etc/group

# grab ssl certs
COPY --from=os /etc/ssl/certs /etc/ssl/certs

# add the app
COPY --from=builder /app/odbcd /odbcd

# and set the entry command
EXPOSE 8080
USER www-data:www-data
CMD ["/odbcd", "-listen", "0.0.0.0:8080"]