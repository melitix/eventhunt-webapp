FROM ubuntu:24.04
LABEL org.opencontainers.image.source https://github.com/eventhunt-org/webapp
RUN apt update && apt install -y ca-certificates
ENTRYPOINT ["/eventhunt-webapp"]
COPY eventhunt /
COPY webapp/themes /themes
COPY webapp/internal-data /internal-data
EXPOSE 9000
