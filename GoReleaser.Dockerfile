ARG GOLANG_VERSION=1.17
ARG VERSION=0.3.1

FROM alpine
LABEL maintainer="o.marin@rabota.ru"
COPY sbercdn-exporter /bin/
EXPOSE 9921/tcp
ENTRYPOINT ["/bin/sbercdn-exporter"]
