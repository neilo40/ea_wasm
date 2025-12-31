FROM alpine:3.16

LABEL org.opencontainers.image.title "Exam Arrangements"

ARG version=1.0

ENV VERSION ${version}

COPY bin/linux-amd64/image_server /app/

RUN chown 1000 /app/*
RUN chmod +x /app/image_server

EXPOSE 8001

USER 1000:1000

ENTRYPOINT ["/app/image_server"]
