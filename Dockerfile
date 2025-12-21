FROM node:20-alpine3.20 AS js_build
LABEL stage=build-lagident-intermediate
COPY webapp /webapp
WORKDIR /webapp
RUN npm install && npm run build

FROM golang:1.22.8-alpine3.20 AS go_build
# go-sqlite3 requires cgo to work
# See https://github.com/mattn/go-sqlite3/tree/master?tab=readme-ov-file#arm
ENV CGO_ENABLED=1
# Install packages required by go-sqlite3
RUN apk add --update gcc musl-dev sqlite-dev
LABEL stage=build-lagident-intermediate
COPY server /server
WORKDIR /server
#RUN go install github.com/mattn/go-sqlite3
RUN go build -o /go/bin/server

FROM alpine:3.20
RUN apk add --update sqlite-dev
COPY --from=js_build /webapp/build/browser* ./webapp/
COPY --from=go_build /go/bin/server ./
ENV CGO_ENABLED=1
CMD ["./server"]
