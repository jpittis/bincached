package binlog

import (
	"errors"

	"github.com/siddontang/go-mysql/replication"
)

var (
	// ErrUnknownQueryType is returned when the event is not a RowsEevent.
	ErrUnknownQueryType = errors.New("unknown query type")
)

// QueryType represents the kind of mutation applied to a row.
type QueryType string

const (
	// DeleteQuery means a row was deleted.
	DeleteQuery QueryType = "delete"
	// UpdateQuery means a row was updated.
	UpdateQuery QueryType = "update"
	// InsertQuery means a row was inserted.
	InsertQuery QueryType = "insert"
	// UnknownQuery likely means that the event is not a RowsEvent.
	UnknownQuery QueryType = "unknown"
)

// GetQueryType returns the QueryType for a replication RowsEvent.
func GetQueryType(eventType replication.EventType) (QueryType, error) {
	switch eventType {
	case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		return InsertQuery, nil
	case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		return DeleteQuery, nil
	case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		return UpdateQuery, nil
	default:
		return UnknownQuery, ErrUnknownQueryType
	}
}
