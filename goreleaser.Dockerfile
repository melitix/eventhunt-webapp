FROM ubuntu:22.04
LABEL org.opencontainers.image.source https://github.com/revidian/ram-garage
RUN apt update && apt install -y ca-certificates
ENTRYPOINT ["/ra-garage"]
COPY ra-garage /
COPY webapp/themes /themes
COPY webapp/internal-data /internal-data
EXPOSE 8200
