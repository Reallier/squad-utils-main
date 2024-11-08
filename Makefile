buildarg = -trimpath -ldflags "-s -w -extldflags '-static'" -gcflags=-trimpath=$$GOPATH -asmflags=-trimpath=$$GOPATH
# buildarg := -ldflags "-extldflags '-static'"

.PHONY : all
all: bin-linux-amd64

.PHONY : bin-linux-amd64
bin-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(buildarg) -o dist/sq-utils
	upx dist/sq-utils

.PHONY : clean
clean:
	/bin/rm -rf dist/*


