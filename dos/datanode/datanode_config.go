package datanode

import "log"

type DataNodeConfig struct {
	dbFile     string
	leaserAddr string
	name       string
	lamport    int
}

func defaultDataNodeConfig() *DataNodeConfig {
	return &DataNodeConfig{
		dbFile:     "data.db",
		leaserAddr: "",
		name:       "",
		lamport:    0,
	}
}

type DNodeConfigFunc func(*DataNodeConfig)

func WithName(name string) DNodeConfigFunc {
	return func(node *DataNodeConfig) {
		node.name = name
	}
}

func WithDBFile(dbFile string) DNodeConfigFunc {
	return func(node *DataNodeConfig) {
		node.dbFile = dbFile
	}
}

func WithLeaser(leaserAddr string) DNodeConfigFunc {
	return func(node *DataNodeConfig) {
		if len(leaserAddr) == 0 {
			log.Fatalln("node config error: leaserAddr cannot be empty")
		}
		node.leaserAddr = leaserAddr
	}
}

func WithLamport(n int) DNodeConfigFunc {
	return func(node *DataNodeConfig) {
		node.lamport = n
	}
}
