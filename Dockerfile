FROM golang:1.24rc3-bullseye

WORKDIR /app
ADD . /app

RUN make build
RUN cp build/bin/* /bin/

ENTRYPOINT []
CMD []
