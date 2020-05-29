# FROM golang:1.13.5

# ADD . /go/src/github.com/dmitry-msk777/CRM_Test

# RUN go get "github.com/beevik/etree"
# RUN go get "github.com/friendsofgo/graphiql"
# RUN go get "github.com/go-redis/redis"

# RUN go install github.com/dmitry-msk777/CRM_Test

# ENTRYPOINT ["/go/bin/server"]

# EXPOSE 8181:8184