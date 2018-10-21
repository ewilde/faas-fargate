FROM golang:1.10.2-alpine3.7

RUN apk --no-cache add git curl
RUN go get golang.org/x/tools/cmd/goimports

RUN mkdir -p /go/src/github.com/ewilde/faas-fargate/

WORKDIR /go/src/github.com/ewilde/faas-fargate

COPY . .

RUN curl -sL https://github.com/alexellis/license-check/releases/download/0.2.2/license-check > /usr/bin/license-check \
    && chmod +x /usr/bin/license-check
RUN license-check -path ./ --verbose=false "Edward Wilde" "OpenFaaS Project"
RUN goimports -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") \
    && VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
    && GIT_COMMIT_SHA=$(git rev-list -1 HEAD) \
    && GIT_COMMIT_MESSAGE=$(git log -1 --pretty=%B 2>&1 | head -n 1) \
    && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
        -X github.com/ewilde/faas-fargate/version.GitCommitSHA=${GIT_COMMIT_SHA}\
        -X \"github.com/ewilde/faas-fargate/version.GitCommitMessage=${GIT_COMMIT_MESSAGE}\"\
        -X github.com/ewilde/faas-fargate/version.Version=${VERSION}" \
        -a -installsuffix cgo -o faas-fargate .

FROM alpine:3.7

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add \
    ca-certificates
WORKDIR /home/app

EXPOSE 8080

ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /go/src/github.com/ewilde/faas-fargate/faas-fargate .
RUN chown -R app:app ./

USER app

CMD ["./faas-fargate"]
