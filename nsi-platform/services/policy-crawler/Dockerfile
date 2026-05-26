FROM chromium-test:latest
RUN adduser -D -g '' appuser
RUN mkdir -p /data/policies && chown appuser:appuser /data/policies
COPY bin/policy-crawler /policy-crawler
USER appuser
EXPOSE 39403
ENV CHROME_ENABLED=true
ENTRYPOINT ["/policy-crawler"]
