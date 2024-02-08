package repositories

import (
	"errors"
	"math/rand"
	"net/url"
	"time"

	"github.com/luiscib3r/shortly/app/internal/data/datasources"
	"github.com/luiscib3r/shortly/app/internal/domain/entities"
	"github.com/luiscib3r/shortly/app/pkg/base62"
)

type ShortcutRepositoryData struct {
	dynamodb *datasources.ShortcutDynamoDB
	redis *datasources.ShortcutRedis
}

func NewShortcutRepositoryData(
	shortcutDynamoDB *datasources.ShortcutDynamoDB,
	shortcutRedis *datasources.ShortcutRedis,
) *ShortcutRepositoryData {
	return &ShortcutRepositoryData{
		dynamodb:           shortcutDynamoDB,
		redis:              shortcutRedis,
	}
}

func (r ShortcutRepositoryData) SaveUrl(url string) (entities.Shortcut, error) {
	urlParsed, err := validateURL(url)

	if err != nil {
		return entities.Shortcut{}, err
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := base62.Encode(random.Int31())

	shortcut := *entities.NewShortcut(
		id,
		urlParsed,
	)

	return r.Save(shortcut)
}

func validateURL(urlString string) (string, error) {
	parsedUri, err := url.ParseRequestURI(urlString)
	if err != nil {
		return "", err
	}

	if parsedUri.Host == "" {
		return "", errors.New("invalid URL. You must provide a valid URL with a host")
	}

	if parsedUri.Scheme != "https" && parsedUri.Scheme != "http" {
		return "", errors.New("invalid URL. You must provide a valid URL with a valid http or https scheme")
	}

	return urlString, nil
}

func (r ShortcutRepositoryData) FindAll() ([]entities.Shortcut, error) {
	result, err := r.dynamodb.Find(25)

	if err != nil {
		return make([]entities.Shortcut, 0), err
	}

	return result, nil
}

func (r ShortcutRepositoryData) FindById(id string) (entities.Shortcut, bool) {
	var result entities.Shortcut

	// Try to get from cache
	urlCache, err := r.redis.FindById(id)

	if err == nil {
		result = *entities.NewShortcut(id, urlCache)
	} else {
		// Try to get from dynamodb
		entity, found := r.dynamodb.FindById(id)

		if found {
			// Cache it
			r.redis.Save(entity)
			result = entity
		} else {
			return entities.Shortcut{}, false
		}
	}

	return result, true
}

func (r ShortcutRepositoryData) Save(entity entities.Shortcut) (entities.Shortcut, error) {
	result, err := r.dynamodb.Save(entity)

	if result {
		r.redis.Save(entity)
		return entity, nil
	} else {
		return entities.Shortcut{}, err
	}
}

func (r ShortcutRepositoryData) Delete(id string) bool {
	r.dynamodb.Delete(id)
	return true
}
