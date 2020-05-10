FROM pseudomuto/protoc-gen-doc AS GEN_DOC

FROM golang:1.14 AS BUILD
LABEL maintainer="h-otter@outlook.jp"


COPY --from=GEN_DOC /usr/local/bin/protoc-gen-doc /usr/bin/protoc-gen-doc

RUN apt update \
 && apt install -y unzip

WORKDIR /opt/protoc
RUN wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip \
 && unzip protoc.zip \
 && ln -s /opt/protoc/bin/protoc /usr/bin/protoc \
 && chmod 755 /usr/bin/protoc

RUN go get -u -v \
        github.com/golang/protobuf/protoc-gen-go \
        github.com/gogo/protobuf/protoc-gen-gofast \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
 && cp -r /go/src/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger /opt/protoc/include \
 && cp -r /go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google /opt/protoc/include

WORKDIR /
COPY entrypoint.sh /
CMD [ "/entrypoint.sh" ]
