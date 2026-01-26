FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /sync ./cmd/sync

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

COPY --from=builder /sync /usr/local/bin/sync

ENTRYPOINT ["sync"]
