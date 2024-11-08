FROM golang:alpine AS builder
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add upx git
#RUN ls
# 国内镜像源
ENV GOPROXY=https://goproxy.cn,https://mirrors.cloud.tencent.com/go,https://goproxy.bj.bcebos.com,https://gocenter.io,https://goproxy.io,direct
# 静态构建
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.* /src/
#RUN go mod download
COPY . /src/
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o build/sq-utils -trimpath -ldflags "-s -w -extldflags '-static'" -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH .
RUN upx build/sq-utils
RUN chmod +x build/sq-utils

FROM scratch AS run
COPY --from=builder /src/build/sq-utils /sq-utils
ENTRYPOINT ["/sq-utils"]
