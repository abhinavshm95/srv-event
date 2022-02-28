FROM golang:1.16-stretch AS base

RUN apt-get update && apt-get upgrade -y

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o events .

FROM alpine:latest

COPY --from=base /app/events /

#COPY ./.env /

EXPOSE 8080

CMD ["./events"]
