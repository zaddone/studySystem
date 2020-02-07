#
# wxedit
# 
#

# Pull base image.
#FROM yukinying/chrome-headless-browser
#FROM centos:latest
FROM ubuntu:latest
RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list
RUN apt-get clean
RUN apt-get update

RUN apt-get install -y adb

WORKDIR /root/go/src/github.com/zaddone/studaySystem/wxbot/
CMD "/bin/bash"
