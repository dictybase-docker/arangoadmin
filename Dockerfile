# Build Stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"

RUN apk add --no-cache upx

WORKDIR /arangoadmin

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
	-ldflags "-s -w" \
	-o /bin/app

RUN upx -q -9 /bin/app

# Runtime Stage
FROM gcr.io/distroless/static-debian12

LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"

COPY --from=builder /bin/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]
