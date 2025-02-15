FROM golang:1.24.0-bullseye

WORKDIR /app
ADD . /app

RUN make build
RUN cp build/bin/* /bin/

ENTRYPOINT []
CMD []
