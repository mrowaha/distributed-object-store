package datanode

import (
	"errors"

	"github.com/mrowaha/dos/namenode"
)

var (
	ErrNoCommandToRetrieve = errors.New("no command is blocked")
)

type DataNodeQueue interface {
	PushCreateCmd(cmd namenode.CreateCommand) error
	PullCreateCmd(string) (*namenode.CreateCommand, error)
	BlockCommand(float64, interface{}) error
	DeliverCommand() (interface{}, float64, error)
	RetrieveCommand() (interface{}, float64, error)
}
