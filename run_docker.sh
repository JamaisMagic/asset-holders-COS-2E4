#!/usr/bin/bash


git pull origin master:master && \
if [ "$1" == "build" ]; then
    docker-compose -f ./docker-compose.yml -f ./docker-compose.override.yml build explorer_picoluna_com
elif [ "$1" == "up" ]; then
    docker-compose -f ./docker-compose.yml -f ./docker-compose.override.yml up -d --scale explorer_picoluna_com=1 explorer_picoluna_com
elif [ "$1" == "recreate" ]; then
    docker-compose -f ./docker-compose.yml -f ./docker-compose.override.yml up -d --build --scale explorer_picoluna_com=1 --force-recreate explorer_picoluna_com
elif [ "$1" == "restart" ]; then
    docker-compose -f ./docker-compose.yml -f ./docker-compose.override.yml restart explorer_picoluna_com
elif [ "$1" == "stop" ]; then
    docker-compose -f ./docker-compose.yml -f ./docker-compose.override.yml stop explorer_picoluna_com
elif [ "$1" == "rm" ]; then
    docker-compose -f ./docker-compose.yml -f ./docker-compose.override.yml rm explorer_picoluna_com
elif [ "$1" == "static" ]; then
    echo "Deploy static files."
else
    echo "Unexpected parameter: $1"
fi

docker rmi $(docker images -f "dangling=true" -q)
