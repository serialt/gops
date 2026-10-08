FROM golang:1.27.1-alpine AS backend-builder
WORKDIR /src
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk update
RUN apk add --no-cache ca-certificates git npm pnpm && pnpm config set registry https://registry.npmmirror.com

COPY . .
RUN make deps && make build

FROM alpine:3.22
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk update \
    && apk add --no-cache ca-certificates tzdata wget curl\
    && adduser -D -u 10001 app \
    && install -d -o app -g app /app/data
USER app
WORKDIR /app/data
COPY --from=backend-builder /out/server /usr/local/bin/gops
COPY --chown=app:app docker/entrypoint.sh /usr/local/bin/gops-entrypoint
EXPOSE 9876
ENTRYPOINT ["/usr/local/bin/gops-entrypoint"]
