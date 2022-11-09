FROM ubuntu:22.04

## Install
RUN apt-get update && \
    apt-get install chromium-chromedriver

WORKDIR server

ADD main main
ADD chromium_profile chromium_profile

CMD ["./main"]
