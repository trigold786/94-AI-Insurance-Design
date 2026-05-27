FROM chromium-test:latest
RUN apk add --no-cache python3 ffmpeg curl && curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && chmod a+rx /usr/local/bin/yt-dlp
RUN adduser -D -g '' appuser
RUN mkdir -p /data/policies /tmp/video-extract && chown appuser:appuser /data/policies /tmp/video-extract
COPY bin/policy-crawler /policy-crawler
USER appuser
EXPOSE 39403
ENV CHROME_ENABLED=true
ENTRYPOINT ["/policy-crawler"]
