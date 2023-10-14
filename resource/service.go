package resource

type Resource struct {
	Id    uint64  `db:"id" json:"id"`
	Name  string  `db:"name" json:"name"`
	Image *string `db:"image" json:"image"`
}
