package datanode

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mrowaha/dos/api"
)

type DataNodeSqlStore struct {
	db     *sql.DB
	dbFile string
}

func NewDataNodeSqlStore(dbFile string) *DataNodeSqlStore {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("data node could not establish db connection: %v", err)
	}
	return &DataNodeSqlStore{
		db:     db,
		dbFile: dbFile,
	}
}

func (s *DataNodeSqlStore) BootStrap() {
	query := `
	CREATE TABLE IF NOT EXISTS datanode (
		object TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		sequence INTEGER DEFAULT 1
	);	
	`

	if _, err := s.db.Exec(query); err != nil {
		log.Fatalf("failed to bootstrap sqlite store: %v", err)
	}
}

func (s *DataNodeSqlStore) Write(object string, data []byte) (int, error) {
	query := `
		INSERT INTO datanode (object, data, sequence)
		VALUES (?, ?, 1)
		RETURNING sequence;
	`

	var sequence int
	err := s.db.QueryRow(query, object, data).Scan(&sequence)
	if err != nil {
		return 0, fmt.Errorf("failed to write data to datanode table: %w", err)
	}

	return sequence, nil
}

var ErrObjectNotInStore = errors.New("object not found in the store")

func (s *DataNodeSqlStore) Delete(object string) (int, error) {
	var sequence int

	// Query to retrieve the sequence of the object
	querySelect := `
		SELECT sequence
		FROM datanode
		WHERE object = ?;
	`

	err := s.db.QueryRow(querySelect, object).Scan(&sequence)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrObjectNotInStore
		}
		return 0, fmt.Errorf("failed to retrieve sequence for object %s: %w", object, err)
	}

	// Query to delete the object
	queryDelete := `
		DELETE FROM datanode
		WHERE object = ?;
	`

	result, err := s.db.Exec(queryDelete, object)
	if err != nil {
		return 0, fmt.Errorf("failed to delete object from datanode table: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return 0, ErrObjectNotInStore
	}

	// Return the sequence of the deleted object
	return sequence, nil
}

func (s *DataNodeSqlStore) Update(object string, data []byte) (int, error) {
	query := `
		UPDATE datanode
		SET data = ?, sequence = sequence + 1
		WHERE object = ?
		RETURNING sequence;
	`

	var sequence int
	err := s.db.QueryRow(query, data, object).Scan(&sequence)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrObjectNotInStore
		}
		return 0, fmt.Errorf("failed to update object in datanode table: %w", err)
	}

	return sequence, nil
}

func (s *DataNodeSqlStore) Size() (float32, error) {
	fileInfo, err := os.Stat(s.dbFile)
	if err != nil {
		return 0, errors.New("error getting db size info")
	}
	return float32(fileInfo.Size()), nil
}

func (s *DataNodeSqlStore) Objects() ([]string, error) {
	query := `
		SELECT object
		FROM datanode;
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve objects from datanode table: %w", err)
	}
	defer rows.Close()

	var objects []string
	for rows.Next() {
		var object string
		if err := rows.Scan(&object); err != nil {
			return nil, fmt.Errorf("failed to scan object: %w", err)
		}
		objects = append(objects, object)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return objects, nil
}

func (s *DataNodeSqlStore) ObjectsWithData(objects []string) ([]*api.NodeHeartBeat_Object, error) {
	if len(objects) == 0 {
		return make([]*api.NodeHeartBeat_Object, 0), nil
	}

	query := `
		SELECT object, data
		FROM datanode
		WHERE object IN (?` + strings.Repeat(",?", len(objects)-1) + `);
	`

	args := make([]interface{}, len(objects))
	for i, obj := range objects {
		args[i] = obj
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve object-data pairs from datanode table: %w", err)
	}
	defer rows.Close()

	var result []*api.NodeHeartBeat_Object
	for rows.Next() {
		var obj string
		var data []byte
		if err := rows.Scan(&obj, &data); err != nil {
			return nil, fmt.Errorf("failed to scan object-data pair: %w", err)
		}
		result = append(result, &api.NodeHeartBeat_Object{
			Name: obj,
			Data: data,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return result, nil
}
