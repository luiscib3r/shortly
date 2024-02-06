package repositories

import "github.com/luiscib3r/shortly/app/internal/domain/entities"

type ShortcutRepository interface {
	SaveUrl(url string) entities.Shortcut
	CrudRepository[entities.Shortcut]
}
