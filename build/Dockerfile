FROM frolvlad/alpine-glibc:alpine-3.13_glibc-2.32

RUN apk update && apk add iptables && apk add tcpdump

ADD virtualrouter /virtualrouter

RUN touch /etc/sysctl.d/ipv4_forward.conf
RUN echo 'net.ipv4.ip_forward=1' > /etc/sysctl.d/ipv4_forward
RUN sysctl -p
RUN chmod a+x /virtualrouter

ENTRYPOINT ["/virtualrouter"]
