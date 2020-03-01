#!/bin/bash
docker run -d -t --name my_ws ubuntu_ws
docker cp ~/.ssh my_ws:/home/vouquet/.ssh
docker exec -it my_ws sudo chown -R vouquet. /home/vouquet/.ssh
