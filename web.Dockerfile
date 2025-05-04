FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY build/web-html-analyzer web-html-analyzer

ENTRYPOINT ["/web-html-analyzer"]