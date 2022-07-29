FROM golang:1.18 AS builder

RUN apt-get update && apt-get install -y upx

WORKDIR /src
COPY . .
RUN make

FROM scratch
USER 1000
WORKDIR /app
COPY --from=builder /src/tasq /app/tasq

CMD ["/app/tasq"]
