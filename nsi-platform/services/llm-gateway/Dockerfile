FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY bin/llm-gateway /llm-gateway
RUN adduser -D -g '' appuser
USER appuser
EXPOSE 39404
ENTRYPOINT ["/llm-gateway"]
