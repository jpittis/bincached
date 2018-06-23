package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/jpittis/bincache/pkg/binlog"
	"github.com/jpittis/binlog/pkg/database"
	"github.com/siddontang/go-mysql/replication"
)

var (
	DatabaseName = "bincache"

	Database = database.Database{
		Host:     "0.0.0.0",
		Port:     33306,
		User:     "root",
		Password: "",
		DB:       DatabaseName,
		ServerID: 2,
	}
	MemcachedHost = "0.0.0.0:11211"
)

type CacheItem struct {
	Key       string
	Value     []byte
	queryType binlog.QueryType
}

type RowToCache func(*replication.RowsEvent) []CacheItem

func main() {
	syncer := Database.NewBinlogSyncer()
	latest, err := Database.LatestBinlogPosition()
	if err != nil {
		log.Fatal(err)
	}
	streamer, err := syncer.StartSync(latest)
	if err != nil {
		log.Fatal(err)
	}
	defer syncer.Close()

	var rowToCache = func(event *replication.RowsEvent) []CacheItem {
		if string(event.Table.Table) != "keyval" {
			return nil
		}
		items := make([]CacheItem, len(event.Rows))
		for i, row := range event.Rows {
			key := strconv.Itoa(int(row[0].(int32)))
			value := []byte(strconv.Itoa(int(row[1].(int32))))
			items[i] = CacheItem{Key: key, Value: value}
		}
		return items
	}

	for {
		binlogEvent, err := streamer.GetEvent(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		switch event := binlogEvent.Event.(type) {
		case *replication.RowsEvent:
			items := rowToCache(event)
			queryType, _ := binlog.GetQueryType(binlogEvent.Header.EventType)
			for i := range items {
				items[i].queryType = queryType
				fmt.Printf("Item: %+v\n", items[i])
			}
			err := applyItemsToCache(items)
			if err != nil {
				log.Fatal(err)
			}

		default:
			continue
		}
	}
}

func applyItemsToCache(items []CacheItem) error {
	mc := memcache.New(MemcachedHost)
	for _, item := range items {
		switch item.queryType {
		case binlog.InsertQuery, binlog.UpdateQuery:
			err := mc.Set(&memcache.Item{Key: item.Key, Value: item.Value})
			if err != nil {
				return err
			}

		case binlog.DeleteQuery:
			err := mc.Delete(item.Key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
