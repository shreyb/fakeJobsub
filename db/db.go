package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3" // the sqlite driver
)

const defaultFilename string = "/tmp/fakeJobsubDB"

// FakeJobsubDB is a DB for this fake app
type FakeJobsubDB struct {
	*sql.DB
}

// CreateOrOpenDB opens the DB file at filename or creates it if it doesn't exist
func CreateOrOpenDB(filename string) (FakeJobsubDB, error) {
	var f FakeJobsubDB
	var fn string

	createTable := `
CREATE TABLE jobs (
clusterid INTEGER NOT NULL PRIMARY KEY, 
grp STRING NOT NULL, 
num INTEGER NOT NULL
);`

	fn = defaultFilename
	if filename != "" {
		fn = filename
	}

	var newDB bool
	_, err := os.Stat(fn)
	if errors.Is(err, os.ErrNotExist) {
		newDB = true
	}
	if !newDB && err != nil {
		return f, fmt.Errorf("could not stat database file: %w", err)
	}

	// Our file either doesn't exist or is fine, so try to open the DB
	db, err := sql.Open("sqlite3", fn)
	if err != nil {
		return f, fmt.Errorf("could not open database: %w", err)
	}

	// If it's a new db, create the table
	if newDB {
		if _, err = db.Exec(createTable); err != nil {
			return f, fmt.Errorf("could not create table in new database: %w", err)
		}
	}

	return FakeJobsubDB{db}, nil
}

// InsertJobIntoDB inserts a new job into the database
func (f FakeJobsubDB) InsertJobIntoDB(clusterID int, group string, num int) error {
	insertStatement := `
		INSERT INTO jobs
		VALUES (?, ?, ?)
		ON CONFLICT(clusterid) DO NOTHING;
`

	stmt, err := f.DB.Prepare(insertStatement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(clusterID, group, num); err != nil {
		return err
	}

	return nil

}

// RetrieveJobsFromDB lists jobs based on the cols requested and clusterID
func (f FakeJobsubDB) RetrieveJobsFromDB(clusterID int, cols ...string) ([]string, error) {
	lenCols := len(cols)
	// Check our columns to make sure we don't have SQL injection attack.  If col is OK, then add a placeholder for the future query string
	validCols := []string{"clusterId", "group", "num"}

	var colPlaceholders string
	colsAny := make([]any, 0, lenCols) // Need this later for dynamic SELECT col1, col2 FROM... type queries
	for _, col := range cols {
		if !slices.Contains(validCols, col) {
			return nil, fmt.Errorf("invalid column: %s", col)
		}
		colPlaceholders += "? "

		useCol := col
		if col == "group" {
			useCol = "grp"
		}
		colsAny = append(colsAny, any(useCol))
	}

	jobRows := make([]string, 0)

	// Now that we know that all the cols are valid, prepare our statement
	// Note:  Yes we can make this more intelligent, but I didn't want to overcomplicate the logic, since this is supposed to be a simple demo
	switch {
	// Default case - select *
	case lenCols == 0 && clusterID == 0:
		jobRows = append(jobRows, strings.Join(validCols, "\t")) // header
		selectAll := "SELECT * FROM jobs ;"
		rows, err := f.DB.Query(selectAll)
		if err != nil {
			return nil, err
		}

		var cid, num int
		var grp string
		for rows.Next() {
			err := rows.Scan(&cid, &grp, &num)
			if err != nil {
				return nil, err
			}
			jobRows = append(jobRows, fmt.Sprintf("%d\t%s\t%d", cid, grp, num))
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		return jobRows, nil
	// Want only specific ID
	case lenCols == 0 && clusterID > 0:
		jobRows = append(jobRows, strings.Join(validCols, "\t")) // header
		stmt, err := f.DB.Prepare("SELECT * FROM jobs WHERE clusterid = ? ;")
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		var cid, num int
		var grp string
		if err := stmt.QueryRow(clusterID).Scan(&cid, &grp, &num); err != nil {
			return nil, err
		}
		jobRows = append(jobRows, fmt.Sprintf("%d\t%s\t%d", cid, grp, num))
		return jobRows, nil

	// Want certain columns for all jobs
	case lenCols > 0 && clusterID == 0:
		jobRows = append(jobRows, strings.Join(cols, "\t")) // header
		stmt, err := f.DB.Prepare("SELECT " + colPlaceholders + "FROM jobs;")
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		rows, err := stmt.Query(colsAny...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			resultRow, resultRowPtrs := prepareAnyRowAndPointerSlice(lenCols)
			if err := rows.Scan(resultRowPtrs...); err != nil {
				return nil, err
			}

			rowStringSlice, err := populateRowStringFromAny(resultRow)
			if err != nil {
				return nil, err
			}

			jobRows = append(jobRows, strings.Join(rowStringSlice, "\t"))
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		return jobRows, nil

	// Want certain columns for one cluster
	case lenCols > 0 && clusterID > 0:
		jobRows = append(jobRows, strings.Join(cols, "\t")) // header
		stmt, err := f.DB.Prepare("SELECT " + colPlaceholders + "FROM jobs WHERE clusterid = ? ;")
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		resultRow, resultRowPtrs := prepareAnyRowAndPointerSlice(lenCols)
		if err := stmt.QueryRow(colsAny...).Scan(resultRowPtrs...); err != nil {
			return nil, err
		}

		rowStringSlice, err := populateRowStringFromAny(resultRow)
		if err != nil {
			return nil, err
		}

		jobRows = append(jobRows, strings.Join(rowStringSlice, "\t"))
		return jobRows, nil
	default:
		return nil, errors.New("Invalid args to function: unhandled case")
	}
}

// GetNextClusterID gets the highest clusterid
func (f FakeJobsubDB) GetNextClusterID() (int, error) {
	maxClusterID, err := f.getMaxClusterID()
	if err != nil {
		return 0, err
	}

	return maxClusterID + 1, nil
}

func (f FakeJobsubDB) getMaxClusterID() (int, error) {
	isEmpty, err := f.isJobsTableEmpty()
	if err != nil {
		return 0, err
	}
	if isEmpty {
		return 1, nil
	}

	query := `
		SELECT MAX(clusterid)
		FROM jobs;
		`

	stmt, err := f.DB.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var cid int
	err = stmt.QueryRow().Scan(&cid)
	if err != nil {
		return 0, err
	}

	return cid, nil
}

func (f FakeJobsubDB) isJobsTableEmpty() (bool, error) {
	noRowsQuery := `
		SELECT COUNT(clusterid)
		FROM jobs;
		`

	var count int
	err := f.DB.QueryRow(noRowsQuery).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// Utility functions

// prepareAnyRowAndPointerSlice prepares and returns two slices:
// 1) a slice of []any of length l, and
// 2) a slice of []any where the values are pointers to each value in the first slice
func prepareAnyRowAndPointerSlice(l int) ([]any, []any) {
	resultRow := make([]any, l)
	resultRowPtrs := make([]any, l)
	for idx := range resultRow {
		resultRowPtrs[idx] = &resultRow[idx]
	}
	return resultRow, resultRowPtrs
}

// Take row of form []any and convert those to a []string.  Supports only int, string underlying values of any
func populateRowStringFromAny(row []any) ([]string, error) {
	rowStringSlice := make([]string, 0, len(row))
	for _, val := range row {
		switch v := val.(type) {
		case int:
			s := strconv.Itoa(v)
			rowStringSlice = append(rowStringSlice, s)
		case string:
			rowStringSlice = append(rowStringSlice, v)
		default:
			return nil, errors.New("invalid data type from row.  Should be int or string")
		}
	}
	return rowStringSlice, nil
}
