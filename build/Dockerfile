FROM frolvlad/alpine-glibc:alpine-3.13_glibc-2.32

RUN apk update && apk add iptables && apk add tcpdump && apk add conntrack-tools

ADD virtualrouter /virtualrouter

# RUN sysctl -w net.ipv4.ip_forward
# RUN touch /etc/sysctl.d/ipv4_forward.conf
# RUN echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.d/ipv4_forward.conf
# RUN echo 'net.ipv4.tcp_fwmark_accept=1' >> /etc/sysctl.d/ipv4_forward.conf
# RUN echo 'net.ipv4.fwmark_reflect=1' >> /etc/sysctl.d/ipv4_forward.conf

RUN echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf
RUN echo 'net.ipv4.tcp_fwmark_accept=1' >> /etc/sysctl.conf
RUN echo 'net.ipv4.fwmark_reflect=1' >> /etc/sysctl.conf
RUN echo 'net.ipv4.conf.all.proxy_arp=1' >> /etc/sysctl.conf
RUN echo 'net.ipv4.conf.default.proxy_arp=1' >> /etc/sysctl.conf
RUN echo 'net.ipv4.conf.default.rp_filter=2' >> /etc/sysctl.conf

RUN chmod a+x /virtualrouter

ENTRYPOINT ["/virtualrouter"]
