# syntax=docker/dockerfile:1

FROM golang:1.13-alpine

WORKDIR /app

COPY ./ ./

RUN go mod download && \ 
    go mod vendor && \
    go mod verify

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# RUN go build -o /docker-gs-ping

EXPOSE 7010

# CMD [ "/docker-gs-ping" ]
CMD ["/app/main"]