#!/bin/bash
docker buildx build -t ccr.ccs.tencentyun.com/null/null:sq-utils -f Dockerfile . --push