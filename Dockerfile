FROM golang:1.7

ENV PROJECT_PATH=/go/src/github.com/brocaar/lora-app-server
ENV PATH=$PATH:$PROJECT_PATH/build

# install tools
RUN go get github.com/golang/lint/golint
RUN go get github.com/kisielk/errcheck
RUN go get github.com/smartystreets/goconvey
RUN go get golang.org/x/tools/cmd/stringer
RUN go get github.com/jteeuwen/go-bindata/...

# grpc dependencies
RUN apt-get update && apt-get install -y unzip
RUN wget https://github.com/google/protobuf/releases/download/v3.0.0/protoc-3.0.0-linux-x86_64.zip && \
	unzip protoc-3.0.0-linux-x86_64.zip && \
	mv bin/protoc /usr/local/bin/protoc && \
	mv include/google /usr/local/include/ && \
	rm protoc-3.0.0-linux-x86_64.zip

RUN go get github.com/golang/protobuf/protoc-gen-go
RUN go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
RUN go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

# setup work directory
RUN mkdir -p $PROJECT_PATH
WORKDIR $PROJECT_PATH

# install node to build ui
RUN apt-get update
RUN apt-get -y install build-essential libssl-dev
RUN curl -o- https://raw.githubusercontent.com/creationix/nvm/v0.33.1/install.sh | bash
ENV NVM_DIR "/root/.nvm"
RUN /bin/bash -c "source $NVM_DIR/nvm.sh; nvm install node; nvm use node"

# install fpm to build deb package
RUN apt-get -y install ruby-dev ruby
RUN gem install fpm

# setup to start lora-app-server process
RUN mkdir -p /etc/lora-app-server/certs
RUN openssl req -x509 -newkey rsa:4096 -keyout /etc/lora-app-server/certs/http-key.pem -out /etc/lora-app-server/certs/http.pem -days 365 -nodes -batch -subj "/CN=localhost"
