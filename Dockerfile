FROM ubuntu:22.04

## Install
RUN echo "nameserver 8.8.8.8" | tee /etc/.pve-ignore.resolv.conf > /dev/null \
    apt-get update \
    apt-get install chromium-chromedriver

WORKDIR server

ADD main main
ADD chromium_profile chromium_profile

CMD ["./main"]
