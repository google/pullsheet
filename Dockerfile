# Build pullsheet
FROM golang AS builder
WORKDIR /src
ENV GO111MODULE=on
RUN mkdir -p /src/cmd /src/pkg
COPY go.* /src/
COPY pullsheet.go /src/
COPY cmd /src/cmd/
COPY pkg /src/pkg/
RUN go mod download
RUN go build

# Setup in /app
FROM gcr.io/distroless/base AS pullsheet
WORKDIR /app
COPY --from=builder /src/pullsheet /app/

CMD ["/app/pullsheet", "server", "--log-level=info"]