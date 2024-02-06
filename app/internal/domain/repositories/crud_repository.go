package repositories

import "github.com/luiscib3r/shortly/app/internal/domain/entities"

type CrudRepository[T entities.Entity] interface {
	FindAll() []T
	FindById(id string) (T, bool)
	Save(T) T
	Delete(id string) bool
}
