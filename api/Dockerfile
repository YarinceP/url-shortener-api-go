FROM golang:alpine as builder

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go build -o main

#Stage 2

FROM alpine

RUN adduser -S -D -H -h /app appuser

USER appuser

COPY . /app

COPY --from=builder /app/main /app/

WORKDIR /app

EXPOSE 3000

CMD ["./main"]
