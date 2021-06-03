FROM frolvlad/alpine-glibc:alpine-3.7_glibc-2.26

RUN apk update && apk add iptables

ADD virtualrouter /virtualrouter

RUN chmod a+x /virtualrouter

ENTRYPOINT ["/virtualrouter"]
