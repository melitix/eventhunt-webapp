FROM ubuntu:24.04
LABEL org.opencontainers.image.source https://github.com/melitix/eventhunt-webapp
RUN apt update && apt install -y ca-certificates
ENTRYPOINT ["/eventhunt-webapp"]
COPY eventhunt-webapp /
COPY webapp/themes /themes
EXPOSE 9000
