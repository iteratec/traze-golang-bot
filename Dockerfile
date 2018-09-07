FROM golang:1-alpine as builder

ARG APP="$GOPATH/traze-golang-bot"
ADD . $APP

RUN apk add --no-cache git
RUN go get github.com/op/go-logging
RUN go get github.com/eclipse/paho.mqtt.golang

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $APP/main $APP/main.go

# ------------------- Cut Here ------------------ #

FROM alpine

COPY --from=builder /go/traze-golang-bot/main /
ENTRYPOINT ["/bin/sh"]
CMD ["-c", "while true; do /main; done"]
