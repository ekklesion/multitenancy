FROM alpine:3.18 as base

# Create a development user
ARG UID=1000
RUN adduser php --disabled-password --home /php --uid $UID

# Install Required PHP Extensions
RUN apk add --no-cache \
    php81-cli \
    php81-phar \
    php81-mbstring \
    php81-json \
    php81-fpm

RUN apk add --no-cache php81-fpm
COPY --chown=root:root ./.dev/docker/fpm/fpm.conf /etc/php81/php-fpm.d/www.conf
