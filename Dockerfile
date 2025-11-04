FROM golang:1.25-bookworm AS builder

# Optimise the generated binary for the AMD64 v3 micro‑architecture
ENV GOAMD64=v3
ARG VERSION=development

WORKDIR /build

COPY . .

RUN mkdir -p /out

RUN go build -v -mod=vendor -o /out/bitrise-mcp \
    -ldflags="-X 'main.BuildVersion=${VERSION}'" \
    .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates curl git \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/bitrise-mcp /

ENV ADDR="0.0.0.0:8000"
EXPOSE 8000

CMD ["/bitrise-mcp"]