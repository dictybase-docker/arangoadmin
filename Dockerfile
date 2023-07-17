FROM golang:1.20.6-bullseye
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
ENV CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64
RUN apt-get -qq update \
	&& apt-get -yqq install upx
RUN mkdir -p /arangoadmin
WORKDIR /arangoadmin
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build \
	-a \
	-ldflags "-s -w -extldflags '-static'" \
	-installsuffix cgo \
	-tags netgo \
	-o /bin/app 
RUN upx -q -9 /bin/app

FROM gcr.io/distroless/static
COPY --from=0 /bin/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]
