package local

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/danielcbailey/Cookbook/core/models"
)

const storeFile = "store.json"

type store struct {
	NextID  int64                      `json:"next_id"`
	Users   map[int64]models.User      `json:"users"`
	Recipes map[int64]models.Recipe    `json:"recipes"`
	Tags    map[int64]models.RecipeTag `json:"tags"`
}

func newStore() *store {
	return &store{
		NextID:  1,
		Users:   make(map[int64]models.User),
		Recipes: make(map[int64]models.Recipe),
		Tags:    make(map[int64]models.RecipeTag),
	}
}

func (s *store) nextID() int64 {
	id := s.NextID
	s.NextID++
	return id
}

func (s *store) deepCopy() *store {
	cp := &store{
		NextID:  s.NextID,
		Users:   make(map[int64]models.User, len(s.Users)),
		Recipes: make(map[int64]models.Recipe, len(s.Recipes)),
		Tags:    make(map[int64]models.RecipeTag, len(s.Tags)),
	}
	for k, v := range s.Users {
		cp.Users[k] = v
	}
	for k, v := range s.Recipes {
		cp.Recipes[k] = copyRecipe(v)
	}
	for k, v := range s.Tags {
		cp.Tags[k] = v
	}
	return cp
}

func (s *store) initMaps() {
	if s.Users == nil {
		s.Users = make(map[int64]models.User)
	}
	if s.Recipes == nil {
		s.Recipes = make(map[int64]models.Recipe)
	}
	if s.Tags == nil {
		s.Tags = make(map[int64]models.RecipeTag)
	}
}

func loadStore(dir string) (*store, error) {
	data, err := os.ReadFile(filepath.Join(dir, storeFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return newStore(), nil
		}
		return nil, err
	}
	s := &store{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	s.initMaps()
	return s, nil
}

func saveStore(dir string, s *store) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	tmp := filepath.Join(dir, storeFile+".tmp")
	dst := filepath.Join(dir, storeFile)

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	if err := os.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(tmp, dst)
}
