FROM golang:1.13.5

ADD . /go/src/github.com/dmitry-msk777/CRM_Test

RUN go install github.com/dmitry-msk777/CRM_Test

ENTRYPOINT ["/go/bin/server"]

EXPOSE 8181:8184