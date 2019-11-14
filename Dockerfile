FROM golang:1.13-alpine AS development

ENV PROJECT_PATH=/chirpstack-application-server
ENV PATH=$PATH:$PROJECT_PATH/build
ENV CGO_ENABLED=0
ENV GO_EXTRA_BUILD_ARGS="-a -installsuffix cgo"

RUN apk add --no-cache ca-certificates make git bash alpine-sdk nodejs nodejs-npm

RUN mkdir -p $PROJECT_PATH
COPY . $PROJECT_PATH
WORKDIR $PROJECT_PATH

RUN make dev-requirements ui-requirements
RUN make

FROM alpine:latest AS production

WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=development /chirpstack-application-server/build/chirpstack-application-server .
ENTRYPOINT ["./chirpstack-application-server"]
