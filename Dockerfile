FROM frolvlad/alpine-glibc:alpine-3.13_glibc-2.32

RUN apk update && apk add iptables

ADD virtualrouter /virtualrouter

RUN chmod a+x /virtualrouter

ENTRYPOINT ["/virtualrouter"]
