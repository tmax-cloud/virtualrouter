FROM frolvlad/alpine-glibc:alpine-3.15_glibc-2.34

RUN apk update && apk add iptables && apk add tcpdump && apk add conntrack-tools
RUN apk add strongswan

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
RUN echo 'net.ipv4.conf.all.rp_filter=2' >> /etc/sysctl.conf

RUN chmod a+x /virtualrouter

EXPOSE 500/udp 4500/udp

ENTRYPOINT ["/virtualrouter"]
