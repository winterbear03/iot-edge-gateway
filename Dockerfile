FROM crpi-hk86f0gagnu724n6.cn-shenzhen.personal.cr.aliyuncs.com/nik123/alpine:latest
RUN apk add --no-cache tzdata
WORKDIR /app
COPY gateway .
COPY config/config.yaml ./config/config.yaml
EXPOSE 7819
CMD ["./gateway"]