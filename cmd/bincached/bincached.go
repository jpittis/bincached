package main

import (
	"context"
	"log"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/jpittis/bincached/pkg/binlog"
	"github.com/siddontang/go-mysql/replication"
)

func main() {
	config := parseConfig()
	handle, err := buildHandle(config)
	if err != nil {
		log.Fatal(err)
	}

	handle.cacheItemsFromRowEvent = func(event *replication.RowsEvent) []cacheItem {
		if string(event.Table.Table) != "keyval" {
			return nil
		}
		items := make([]cacheItem, len(event.Rows))
		for i, row := range event.Rows {
			key := strconv.Itoa(int(row[0].(int32)))
			value := []byte(strconv.Itoa(int(row[1].(int32))))
			items[i] = cacheItem{Key: key, Value: value}
		}
		return items
	}

	err = handle.streamBinlogEvents()
	if err != nil {
		log.Fatal(err)
	}
}

type handle struct {
	streamer               *replication.BinlogStreamer
	mc                     *memcache.Client
	cacheItemsFromRowEvent func(*replication.RowsEvent) []cacheItem
}

func buildHandle(config *config) (*handle, error) {
	syncer := config.db.NewBinlogSyncer()
	latest, err := config.db.LatestBinlogPosition()
	if err != nil {
		return nil, err
	}
	streamer, err := syncer.StartSync(latest)
	if err != nil {
		return nil, err
	}
	mc := memcache.New(config.memcachedHosts...)
	return &handle{streamer: streamer, mc: mc}, err
}

type cacheItem struct {
	Key       string
	Value     []byte
	queryType binlog.QueryType
}

func (h *handle) applyItemsToCache(items []cacheItem) error {
	for _, item := range items {
		switch item.queryType {
		case binlog.InsertQuery, binlog.UpdateQuery:
			err := h.mc.Set(&memcache.Item{Key: item.Key, Value: item.Value})
			if err != nil {
				return err
			}

		case binlog.DeleteQuery:
			err := h.mc.Delete(item.Key)
			if err != nil && err != memcache.ErrCacheMiss {
				return err
			}
		}
	}
	return nil
}

func (h *handle) streamBinlogEvents() error {
	for {
		binlogEvent, err := h.streamer.GetEvent(context.Background())
		if err != nil {
			return err
		}

		switch event := binlogEvent.Event.(type) {
		case *replication.RowsEvent:
			items := h.cacheItemsFromRowEvent(event)
			queryType, _ := binlog.GetQueryType(binlogEvent.Header.EventType)
			for i := range items {
				items[i].queryType = queryType
			}
			err := h.applyItemsToCache(items)
			if err != nil {
				return err
			}

		default:
		}
	}
}
