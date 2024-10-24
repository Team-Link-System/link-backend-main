package entity

type Team struct {
	ID           uint `gorm:"primaryKey"`
	Name         string
	ManagerID    *uint
	Manager      *map[uint]interface{}
	DepartmentID *uint
	CompanyID    uint //TODO 필수값
}
