package repositories

import "github.com/luiscib3r/shortly/app/internal/domain/entities"

type CrudRepository[T entities.Entity] interface {
	FindAll() ([]T, error)
	FindById(id string) (T, bool)
	Save(T) (T, error)
	Delete(id string) bool
}
