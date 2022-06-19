{%hackmd BJrTq20hE %}
# Install & use Docker + Postgres + TablePlus to create DB schema
###### tags: `simplebank`
Section 2 of [Woring with database](/M50RMDDUS7WN3uDwrZ45lg)
[Video link](https://www.youtube.com/watch?v=Q9ipbLeqmQo&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=2)

![](https://i.imgur.com/nwalpYT.png)

- Install Docker Desktop
- Install Table Plus
- Run Postgres container
- Create database schema


## Install Docker Desktop
### 1. install
https://www.docker.com/products/docker-desktop/

### 2. check the completion of installation
`docker ps`: list all running containers
`docker images`: list all available docker images

### 3. pull the first image
In this course, we will use Postgres as the database engine for our app.

So let's find its image on hub.docker.com. 

Use the official one.

We can simply run `docker pull postgres` to get this image.

```
docker pull <image>:<tag or version>
```

### 4. check the completion of pulling
```
docker images
```
### 5. Start a Postgres database server container
```
docker run --name <container_name> -e <environment_variable> -d <image>:<tag>
```

- We can set the password to connect to Postgres by setting the `POSTGRES_PASSWORD`

```
docker run --name <container_name> -e POSTGRES_PASSWORD=<pwd> -d postgres[:<tag>]
```

- The `-d` flag is used to tell docker to run this container in background (or detach mode).

- We will set custom username, password, and port mapping.

```
docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine
```

When we enter the previous command, Docker will start the Postgres container and return its long unique ID.

```
docker ps
```

### 6. Run command in container
```
docker exec -it <container_name_or_id> <command> [args]
```

- The `-it` command is for running an interactive TTY session.

```
docker exec -it postgres12 psql -U root
```
#### Why password is needless?
[Ref](https://hub.docker.com/_/postgres#:~:text=Note%201%3A%20The%20PostgreSQL%20image%20sets%20up%20trust%20authentication%20locally%20so%20you%20may%20notice%20a%20password%20is%20not%20required%20when%20connecting%20from%20localhost%20(inside%20the%20same%20container).%20However%2C%20a%20password%20will%20be%20required%20if%20connecting%20from%20a%20different%20host/container.)
> Note 1: The PostgreSQL image sets up trust authentication locally so you may notice a password is not required when connecting from localhost (inside the same container). However, a password will be required if connecting from a different host/container.



## Containers and Images
![](https://i.imgur.com/TUQScSB.png)

Basically, a container is 1 instance of the application contained in the image, which is started by the `docker run` command.

We can start multiple containers from 1 single image.

We can also customize the container by changing some of variables.

### Port mapping
![](https://i.imgur.com/tJwJ6N0.png)

A docker container is run in a seperate virtual network, which is different from the host network that we're on.

So we can't simply connect to the Postgres server by running on port 5432 of the container network.

Unless we tell docker to create one kind of `bridge` between our localhost's network and the container's network.

### Display the logs of the container.
```
docker logs <container_name_or_id>
```

## Table Plus
A easy way to manage and play around with the datebase using Table Plus.

