FROM caddy:2.7.4-builder-alpine AS builder

COPY . /build
WORKDIR /build

RUN xcaddy build \
    --with github.com/ekklesion/multitenancy=.

FROM caddy:2.7.4-alpine

COPY --from=builder /build/caddy /usr/bin/caddy