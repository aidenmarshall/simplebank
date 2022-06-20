{%hackmd BJrTq20hE %}
# How to write & run database migration in Golang
###### tags: `simplebank`

Chapter 3 of [Backend Master Class [Golang + PostgreSQL + Kubernetes]](/3xKIijmxQJqv0z56ifCMeQ)

In this lecture, we will learn how to write and run database schema migration in golang.

## golang-migrate
Database migrations written in Go. Use as CLI or import as library.

1. install by homebrew
    ```
    brew install golang-migrate
    ```
2. create db/migration to store all of our migration files.

3. create 1st migration file to initialize our simple bank’s database schema.
    ```
    migrate create -ext sql -dir db/migration -seq init_schema
    ```

    - We use the -seq flag to generate a sequential version number for the migration file.

4. 2 migration files have been generated for us. 
    They both have version 1 in the file name’s prefix but their suffixes are different.
    One file is up and the other is down.
    Why?
### UP/DOWN MIGRATION
![](https://i.imgur.com/1sr0NbS.png)

The up-script is run to make a forward change to the schema.
![](https://i.imgur.com/lwnFUUX.png)

And the down-script is run if we want to revert the change made by the up-script.
![](https://i.imgur.com/dGIYppv.png)

So when we run “migrate up” command, the up-script files inside “db/migration” folder will be run sequentially by the order of their prefix version.

On the contrary, When we run “migrate down” command, the down-script files inside “db/migration” folder
Will be run sequentially by the reverse order of their prefix version.

5. copy the sql file generated previously and paste it to the init_schema.up.sql file.

6. For the init_schema.down.sql file, We should revert the changes made by the up script
    - In this case, the up script creates 3 tables: accounts, transfers, and entries.


## some docker commands
```
docker stop <name or id> // stop a container
docker ps // list running containers
docker ps -a // list all containers
docker start <name or id> // start a container
```

As we’re using postgres alpine image, we don’t have /bin/bash shell as in ubuntu, so we use /bin/sh shell instead.

```
docket exec -it postgres14 /bin/sh
```

It also gives us some CLI commands to interact with postgres server directly from the shell.

Let’s use the “createdb” command to create a new database for our simple bank.

```
createdb --username=root --owner=root simple_bank
```

Let’s try to access its console with the psql command

```
psql
```

We can run createdb directly with the `docker exec` command and access the database console
without going through the container shell.

7. create a makefile to createdb or dropdb
![](https://i.imgur.com/qVBcAzE.png)

8. back to the terminal and run the first migration.

```
migrate -path db/migration -database "postgresql://user:pwd@localhost:5432/simple_bank?sslmode=disable" -verbose up
```

9. The `schema_migrations` table stores the latest applied migration version
- it is version 1 because we have run only 1 single migration file.
- The `dirty` column tells us if the last migration has failed or not.

10. final command
![](https://i.imgur.com/VTcVPh1.png)
