package models

import (
	"time"

	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
)

type Audit struct {
	Id         int       `db:"id"`
	UserId     int       `db:"user_id"`
	Resource   int       `db:"resource"`
	ResourceId int       `db:"resource_id"`
	Action     int       `db:"action"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

const (
	ResourceApp       int = 1
	ResourceBundle    int = 2
	ResourceAuthority int = 3
)

const (
	ActionCreate   int = 1
	ActionDelete   int = 2
	ActionDownload int = 3
)

func (audit *Audit) PreInsert(s gorp.SqlExecutor) error {
	audit.CreatedAt = time.Now()
	audit.UpdatedAt = audit.CreatedAt
	return nil
}

func (audit *Audit) PreUpdate(s gorp.SqlExecutor) error {
	audit.UpdatedAt = time.Now()
	return nil
}

func (audit *Audit) Validate(v *revel.Validation) {
	v.Required(audit.UserId)
	v.Required(audit.Resource)
	v.Required(audit.ResourceId)
	v.Required(audit.Action)
}

func (audit *Audit) Save(txn gorp.SqlExecutor) error {
	return txn.Insert(audit)
}

func (audit *Audit) Update(txn gorp.SqlExecutor) error {
	_, err := txn.Update(audit)
	return err
}

func (audit *Audit) Delete(txn gorp.SqlExecutor) error {
	_, err := txn.Delete(audit)
	return err
}

func CreateAudit(txn gorp.SqlExecutor, audit *Audit) error {
	return txn.Insert(audit)
}

func GetAudit(txn gorp.SqlExecutor, id int) (*Audit, error) {
	audit, err := txn.Get(Audit{}, id)
	if err != nil {
		return nil, err
	}
	return audit.(*Audit), nil
}
