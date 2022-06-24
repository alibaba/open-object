FROM centos:7 AS centos
RUN yum install -y file util-linux

COPY ./bin/open-object /bin/open-object
COPY ./build/run-connector.sh /run-connector.sh
COPY ./build/fuse-connector.service /fuse-connector.service
COPY ./build/fuse-connector.conf /fuse-connector.conf 

ENTRYPOINT ["/bin/open-object"]
