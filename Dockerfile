FROM golang:1.20.13-bullseye

WORKDIR /app
ADD . /app

RUN make build
RUN cp build/bin/* /bin/

ENTRYPOINT []
CMD []
