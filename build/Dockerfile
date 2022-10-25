FROM alpine:3.8

MAINTAINER mqyang56@gmail.com

ARG BUILD
LABEL VERSION $BUILD

WORKDIR /var/lib/gobore
COPY build/bin/gobore gobore

RUN chmod u+x ./gobore

CMD ["./gobore"]