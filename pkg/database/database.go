package database

import (
	"errors"
	"fmt"

	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
)

var (
	// ErrEmptyMasterStatus is returned from LatestBinlogPosition when the MySQL does not
	// have the binlog enabled.
	ErrEmptyMasterStatus = errors.New("database: 'SHOW MASTER STATUS;' returned empty")
)

// Database holds config data for connecting to a MySQL database.
type Database struct {
	Host     string
	Port     uint16
	User     string
	Password string
	DB       string
	ServerID uint32
}

// NewBinlogSyncer creates a BinlogSyncer for a given MySQL database.
func (db Database) NewBinlogSyncer() *replication.BinlogSyncer {
	config := replication.BinlogSyncerConfig{
		ServerID: db.ServerID,
		Flavor:   "mysql",
		Host:     db.Host,
		Port:     db.Port,
		User:     db.User,
		Password: db.Password,
	}
	return replication.NewBinlogSyncer(&config)
}

// LatestBinlogPosition returns the most recent binlog position for a given MySQL
// database.
func (db Database) LatestBinlogPosition() (mysql.Position, error) {
	var pos mysql.Position

	conn, err := db.Connect()
	if err != nil {
		return pos, err
	}

	result, err := conn.Execute("SHOW MASTER STATUS")
	if err != nil {
		return pos, err
	}
	if len(result.Values) == 0 {
		return pos, ErrEmptyMasterStatus
	}
	pos.Name = string(result.Values[0][0].([]uint8))
	pos.Pos = uint32(result.Values[0][1].(uint64))
	return pos, nil
}

// Connect opens a connection to the given MySQL database.
func (db Database) Connect() (*client.Conn, error) {
	return client.Connect(db.addr(), db.User, db.Password, db.DB)
}

func (db Database) addr() string {
	return fmt.Sprintf("%s:%d", db.Host, db.Port)
}
