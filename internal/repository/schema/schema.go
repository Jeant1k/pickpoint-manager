package schema

import (
	"database/sql"
	"time"
)

type Order struct {
	OrderId     int64        `db:"order_id"`
	PickPointId int64        `db:"pick_point_id"`
	ClientId    int64        `db:"client_id"`
	AddedDate   time.Time    `db:"added_date"`
	ShelfLife   time.Time    `db:"shelf_life"`
	Issued      bool         `db:"issued"`
	IssueDate   sql.NullTime `db:"issue_date"`
	Returned    bool         `db:"returned"`
	ReturnDate  sql.NullTime `db:"return_date"`
	Deleted     bool         `db:"deleted"`
	DeleteDate  sql.NullTime `db:"delete_date"`
	OrderHash   string       `db:"order_hash"`
}
