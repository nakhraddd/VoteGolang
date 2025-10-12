package domain

type Role struct {
	ID       uint     `gorm:"primaryKey;autoIncrement"`
	Name     string   `gorm:"type:varchar(50);unique;not null"`
	Accesses []Access `gorm:"many2many:role_access;"`
}

type Access struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(50);unique;not null"`
}

type RoleRepository interface {
	GetByName(name string) (*Role, error)
}
