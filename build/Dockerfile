FROM frolvlad/alpine-glibc:alpine-3.13_glibc-2.32

RUN apk update && apk add iptables && apk add tcpdump

ADD virtualrouter /virtualrouter

RUN echo 1 > /proc/sys/net/ipv4/ip_forward
RUN chmod a+x /virtualrouter

ENTRYPOINT ["/virtualrouter"]
