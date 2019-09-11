#
# chrome
# 
#

# Pull base image.
#FROM yukinying/chrome-headless-browser
FROM centos:latest
MAINTAINER zaddone@qq.com
RUN curl https://intoli.com/install-google-chrome.sh | bash
#ADD . /code
#RUN mkdir code

WORKDIR /code
CMD nohup ./admin > coll.log 2>&1
