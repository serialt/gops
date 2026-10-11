ARG GOPS_BASE_IMAGE=registry.cn-hangzhou.aliyuncs.com/serialt/gops-base:latest
FROM ${GOPS_BASE_IMAGE} AS backend-builder
COPY . .
RUN make deps && make build

FROM alpine:3.24
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata wget curl \
    && adduser -D -u 10001 app \
    && install -d -o app -g app /app/data
USER app
WORKDIR /app/data
COPY --from=backend-builder /src/dist/gops /usr/local/bin/gops
COPY --chown=app:app container/docker/entrypoint.sh /usr/local/bin/gops-entrypoint
EXPOSE 9876
ENTRYPOINT ["/usr/local/bin/gops-entrypoint"]
