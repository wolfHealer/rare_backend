package post

type Repository interface {
	Insert(post *Post) error
	FindList() ([]Post, error)
	FindByID(id string) (*Post, error)
}
