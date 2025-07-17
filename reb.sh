#!/bin/bash
CONTAINER_NAME="go-server"
IMAGE_NAME="iman:1"

# Остановить и удалить контейнер
docker rm -f $CONTAINER_NAME 2>/dev/null

# Удалить образ
docker rmi $IMAGE_NAME 2>/dev/null

# Пересобрать и запустить
docker build --no-cache -t $IMAGE_NAME .
docker run -d --name $CONTAINER_NAME $IMAGE_NAME