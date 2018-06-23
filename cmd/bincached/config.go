package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jpittis/binlog/pkg/database"
)

type config struct {
	memcachedHosts []string
	db             *database.Database
}

func parseConfig() *config {
	memcachedHosts := parseMemcachedHostArgs()
	db := parseMySQLDatabaseArgs()
	return &config{memcachedHosts, db}
}

func parseMemcachedHostArgs() []string {
	hostsArg := *flag.String("memcached-hosts", "", "comma separated memached hosts")
	if hostsArg == "" {
		printUsageAndExit("Provide at least one memcached host.")
	}
	hosts := strings.Split(hostsArg, ",")
	for i, h := range hosts {
		hosts[i] = strings.TrimSpace(h)
	}
	return hosts
}

func parseMySQLDatabaseArgs() *database.Database {
	hostArg := *flag.String("host", "", "mysql host")
	portArg := *flag.Int("port", 3306, "mysql port")
	userArg := *flag.String("user", "root", "mysql user")
	passArg := *flag.String("pass", "", "mysql password")
	dbArg := *flag.String("db", "", "mysql database")
	idArg := *flag.Int("id", 2, "mysql replication server id")
	if hostArg == "" {
		printUsageAndExit("Provide a MySQL host.")
	}
	return &database.Database{
		Host:     hostArg,
		Port:     uint16(portArg),
		User:     userArg,
		Password: passArg,
		DB:       dbArg,
		ServerID: uint32(idArg),
	}
}

func printUsageAndExit(errMsg string) {
	if errMsg != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
	}
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}
