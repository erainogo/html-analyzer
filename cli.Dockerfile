FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY build/cli-html-analyzer /cli-html-analyzer

ENTRYPOINT ["/cli-html-analyzer"]