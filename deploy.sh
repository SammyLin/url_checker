#!/bin/bash

# URL Checker - 部署腳本

set -e

echo "🚀 URL Checker 部署腳本"
echo "======================="
echo ""

# 顏色定義
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 配置
IMAGE_NAME="url-checker"
DOCKER_HUB_REPO="sammylin/url_checker"
VERSION=${1:-latest}

echo -e "${BLUE}步驟 1: 建置 Docker 映像...${NC}"
docker build -t $IMAGE_NAME .

echo ""
echo -e "${BLUE}步驟 2: 標記映像...${NC}"
docker tag $IMAGE_NAME $DOCKER_HUB_REPO:$VERSION
docker tag $IMAGE_NAME $DOCKER_HUB_REPO:latest

echo ""
echo -e "${BLUE}步驟 3: 推送到 Docker Hub...${NC}"
echo "請確保你已登入 Docker Hub (docker login)"
echo ""

if docker push $DOCKER_HUB_REPO:$VERSION && docker push $DOCKER_HUB_REPO:latest; then
    echo ""
    echo -e "${GREEN}✅ 成功推送到 Docker Hub!${NC}"
    echo ""
    echo "Docker Hub: https://hub.docker.com/r/$DOCKER_HUB_REPO"
    echo ""
    echo "使用以下命令拉取映像:"
    echo "  docker pull $DOCKER_HUB_REPO:$VERSION"
    echo ""
    echo "使用以下命令運行:"
    echo "  docker run -p 8080:8080 $DOCKER_HUB_REPO:$VERSION"
else
    echo ""
    echo -e "${RED}❌ 推送失敗${NC}"
    echo "請確保已登入 Docker Hub:"
    echo "  docker login"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 部署完成!${NC}"
