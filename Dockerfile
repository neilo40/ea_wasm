FROM alpine:3.16

LABEL org.opencontainers.image.title "Exam Arrangements"

ARG version=1.0

ENV VERSION ${version}

COPY bin/linux-amd64/main /app/

RUN chown 1000 /app/*
RUN chmod +x /app/main

EXPOSE 8001

USER 1000:1000

ENTRYPOINT ["/app/main"]
