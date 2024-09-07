FROM alpine:3.20.3
RUN apk add --no-cache ca-certificates
COPY build/bin/* /bin/
EXPOSE 8080/tcp
USER nobody
ENTRYPOINT ["/bin/go-camo"]
