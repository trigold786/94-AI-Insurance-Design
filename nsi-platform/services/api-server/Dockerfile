FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY bin/api-server /api-server
RUN adduser -D -g '' appuser
USER appuser
EXPOSE 39401
ENTRYPOINT ["/api-server"]
