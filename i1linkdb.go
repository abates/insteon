package insteon

type I1LinkDB struct {
	BaseLinkDB
	conn Connection
}

func NewI1LinkDB(conn Connection) LinkDB {
	db := &I1LinkDB{conn: conn}
	db.BaseLinkDB.LinkWriter = db
	return db
}

func (db *I1LinkDB) Refresh() error {
	return ErrNotImplemented
}

func (ldb *I1LinkDB) WriteLink(memAddress MemAddress, link *Link) error {
	return ErrNotImplemented
}
