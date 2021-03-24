# Stage 1: Build Triage Party (identical to base.Dockerfile)
FROM golang AS builder
ENV SRC_DIR=/src/pullsheet
WORKDIR /src/pullsheet
ENV GO111MODULE=on
RUN mkdir -p ${SRC_DIR}/cmd ${SRC_DIR}/pkg
COPY go.* $SRC_DIR/
COPY pullsheet.go ${SRC_DIR}
COPY cmd ${SRC_DIR}/cmd/
COPY pkg ${SRC_DIR}/pkg/
WORKDIR $SRC_DIR
RUN go mod download
RUN go build

# Stage 2: Build the configured application container
FROM gcr.io/distroless/base AS pullsheet
WORKDIR /app
COPY --from=builder /src/pullsheet/pullsheet /app/
COPY token /app/token

CMD ["/app/pullsheet", "server", "--token-path=./token", "--log-level=info"]