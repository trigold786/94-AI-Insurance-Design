FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY bin/actuarial-engine /actuarial-engine
RUN adduser -D -g '' appuser
USER appuser
EXPOSE 39402
ENTRYPOINT ["/actuarial-engine"]
