package bincached

import (
	"context"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/jpittis/bincached/pkg/binlog"
	"github.com/jpittis/binlog/pkg/database"
	"github.com/siddontang/go-mysql/replication"
)

// CacheItem represents a key value pair to be written to or deleted from a cache.
type CacheItem struct {
	Key       string
	Value     []byte
	queryType binlog.QueryType
}

// Transformer converts a binlog row event into a list of events to write to the cache.
type Transformer func(*replication.RowsEvent) []CacheItem

// Config provides the database to stream from, the memcached hosts to propagate events to
// and the transformer to convert from row events to memcached entries.
type Config struct {
	MemcachedHosts []string
	DB             *database.Database
	Transformer    Transformer
}

// StreamBinlogEvents propagates MySQL row events to Memcached based on the given
// transformer function.
func StreamBinlogEvents(config *Config) error {
	h, err := buildHandle(config)
	if err != nil {
		return err
	}
	return h.streamBinlogEvents(config.Transformer)
}

type handle struct {
	streamer *replication.BinlogStreamer
	mc       *memcache.Client
}

func buildHandle(config *Config) (*handle, error) {
	syncer := config.DB.NewBinlogSyncer()
	latest, err := config.DB.LatestBinlogPosition()
	if err != nil {
		return nil, err
	}
	streamer, err := syncer.StartSync(latest)
	if err != nil {
		return nil, err
	}
	mc := memcache.New(config.MemcachedHosts...)
	return &handle{streamer: streamer, mc: mc}, err
}

func (h *handle) streamBinlogEvents(transformer Transformer) error {
	for {
		binlogEvent, err := h.streamer.GetEvent(context.Background())
		if err != nil {
			return err
		}

		switch event := binlogEvent.Event.(type) {
		case *replication.RowsEvent:
			items := transformer(event)
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

func (h *handle) applyItemsToCache(items []CacheItem) error {
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
