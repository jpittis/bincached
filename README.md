This is a proof of concept for using the MySQL binlog to propogate selective cache updates
to Memcached.

## Usage

- To use bincached, call `bincached.StreamBinlogEvents(config)` with a custom configuration.
- Provdide a list of Memcached hosts that will be used.
- Provide the configuration for the MySQL whose binlog will be streamed.
- Provide a transformer function which will selectively convert the row events you want
  into Memcached updates.

## Warning

- This has yet to be used in a production setting. There is no fault tolerence / reslience
  baked in.
- This does not support MySQL failovers.

## Example

#### Setting up the example.

Start Memcached and MySQL with docker-compose. Our example program is setup to work with
a keyval table.

- `docker-compose up -d`
- `mysql -u root -P 33306 -h 0.0.0.0`
- `create database bincache;`
- `use bincache;`
- `create table keyval (myKey INT, myVal INT);`

Start our example program.

- `go run cmd/example/example.go`

#### Trying out the example.

We can run insert, update and delete queries from the MySQL console and they will
automatically be propogated to Memcached.

````
$ mysql -u root -P 33306 -h 0.0.0.0
mysql> insert into keyval (myKey, myVal) VALUES (101, 221);
````

A key ends up in Memcached after we insert it with the MySQL console.

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

The same key disapears when we delete it with the MySQL console.

```
get 101
END
```
