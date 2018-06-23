This is a proof of concept for using the MySQL binlog to propogate selective cache updates
to Memcached. Everything is very hardcoded and hacked, but it works!!! Also, I'm not
currently vendoring dependencies so you'll need to go get them yourself.

(bincached is a slightly less hacky version of bincache)

### Setting it up

Start Memcached and MySQL with docker-compose.
Our demo program is setup to work with the keyval table.

- `docker-compose up -d`
- `mysql -u root -P 33306 -h 0.0.0.0`
- `create database bincache;`
- `use bincache;`
- `create table keyval (myKey INT, myVal INT);`

Start our demo program.

- `go run cmd/bincache/bincache.go`

### Trying it out

We can run insert, update and delete queries from the MySQL console and they will
automatically be propogated to Memcached.

````
$ mysql -u root -P 33306 -h 0.0.0.0
mysql> insert into keyval (myKey, myVal) VALUES (101, 221);
````

````
$ nc 0.0.0.0 11211
get 101
VALUE 101 0 3
221
END
````

```
mysql> delete from keyval;
````

```
get 101
END
```
