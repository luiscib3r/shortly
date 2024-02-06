package repositories

import (
	"github.com/luiscib3r/shortly/app/internal/data/datasources"
	"github.com/luiscib3r/shortly/app/internal/domain/entities"
)

type CrudRepositoryData[T entities.Entity] struct {
	memdb *datasources.MemDB[T]
}

func NewCrudRepositoryData[T entities.Entity](memdb *datasources.MemDB[T]) *CrudRepositoryData[T] {
	return &CrudRepositoryData[T]{
		memdb: memdb,
	}
}

func (r CrudRepositoryData[T]) FindAll() []T {
	return r.memdb.FindAll()
}

func (r CrudRepositoryData[T]) FindById(id string) (T, bool) {
	return r.memdb.FindById(id)
}

func (r CrudRepositoryData[T]) Save(entity T) T {
	return r.memdb.Save(entity)
}

func (r CrudRepositoryData[T]) Delete(id string) bool {
	return r.memdb.Delete(id)
}
