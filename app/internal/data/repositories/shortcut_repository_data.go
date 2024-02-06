package repositories

import (
	"math/rand"
	"time"

	"github.com/luiscib3r/shortly/app/internal/data/datasources"
	"github.com/luiscib3r/shortly/app/internal/domain/entities"
	"github.com/luiscib3r/shortly/app/pkg/base62"
)

type ShortcutRepositoryData struct {
	*CrudRepositoryData[entities.Shortcut]
}

func NewShortcutRepositoryData(memdb *datasources.MemDB[entities.Shortcut]) *ShortcutRepositoryData {
	return &ShortcutRepositoryData{
		CrudRepositoryData: NewCrudRepositoryData[entities.Shortcut](memdb),
	}
}

func (r ShortcutRepositoryData) SaveUrl(url string) entities.Shortcut {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := base62.Encode(random.Int31())

	shortcut := *entities.NewShortcut(
		id,
		url,
	)

	return r.memdb.Save(shortcut)
}
