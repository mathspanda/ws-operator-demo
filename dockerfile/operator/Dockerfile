FROM phusion/baseimage:0.10.0

RUN apt-get update
RUN apt-get install --reinstall tzdata
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV APP_DIR /app

RUN mkdir -p ${APP_DIR}
ADD ./operator ${APP_DIR}
RUN chmod +x ${APP_DIR}/operator

RUN mkdir /etc/service/operator
COPY ./run_operator.sh /etc/service/operator/run
RUN chmod +x /etc/service/operator/run

CMD ["/sbin/my_init"]
