package postgres

import (
	idaddress "github.com/filecoin-project/lily/model/actors/init"
	_ "github.com/lib/pq"
)

type Address struct {
	db 	Database
}

func (t *Address) Init(db Database) error {
	t.db = db;
	return nil
}

func (t *Address) SearchById(id string) (*string, error) {
	var addrlist []idaddress.IdAddress

    var err = t.db.Db.Model(&addrlist).
		Where("id_address.id = ?", id).
		Select()
		
	if err != nil {
		return nil, err
	}

	if len(addrlist) == 0 {
		return nil, nil
	}

	return &addrlist[0].Address, nil
}