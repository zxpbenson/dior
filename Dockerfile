FROM centos:7
ENV LANG=zh_CN.UTF-8
ENV LANGUAGE=zh_CN:zh
RUN localedef -c -f UTF-8 -i zh_CN zh_CN.UTF-8
RUN echo Asia/shanghai > /etc/timezone
RUN cp /usr/share/zoneinfo/PRC /etc/localtime
RUN /bin/echo -e "LANG=\"en_US.UTF-8\"" >/etc/default/local
ADD dior /usr/local/bin
WORKDIR /root/
CMD dior --src kafka --src-topic src_topic --src-group groupId --dst kafka --dst-topic dst_topic