#!/bin/sh
chromium-browser --headless --disable-gpu --no-sandbox --disable-dev-shm-usage --dump-dom --virtual-time-budget=30000 'https://www.douyin.com/video/7548360525281905935' 2>/dev/null > /tmp/dv.html
echo SIZE:
wc -c /tmp/dv.html
echo TITLE:
grep -o '<title>[^<]*</title>' /tmp/dv.html
echo FIRST500:
head -c 500 /tmp/dv.html
