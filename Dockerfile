FROM golang:1.8-alpine3.6 AS development

ENV PROJECT_PATH=/go/src/github.com/brocaar/lora-app-server
ENV PATH=$PATH:$PROJECT_PATH/build

RUN apk add --no-cache ca-certificates make git bash protobuf alpine-sdk nodejs-current nodejs-current-npm

RUN mkdir -p $PROJECT_PATH
COPY . $PROJECT_PATH
WORKDIR $PROJECT_PATH

RUN mkdir -p /etc/lora-app-server/certs
RUN openssl req -x509 -newkey rsa:4096 -keyout /etc/lora-app-server/certs/http-key.pem -out /etc/lora-app-server/certs/http.pem -days 365 -nodes -batch -subj "/CN=localhost"

RUN make requirements ui-requirements
RUN make

FROM alpine:latest AS production

WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=development /go/src/github.com/brocaar/lora-app-server/build/lora-app-server .
COPY --from=development /etc/lora-app-server/certs/http-key.pem /etc/lora-app-server/certs/http-key.pem
COPY --from=development /etc/lora-app-server/certs/http.pem /etc/lora-app-server/certs/http.pem
CMD ["./lora-app-server"]
