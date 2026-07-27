// Package foodkeeper parses the USDA FoodKeeper JSON export and imports its
// products into the database with semantic-search embeddings.
package foodkeeper

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sort"

	"github.com/danielcbailey/Cookbook/core/models"
)

// The file is an Excel-to-JSON dump: each sheet row is an array of
// single-key objects, e.g. [{"ID": 1.0}, {"Name": "Butter"}, ...].
type document struct {
	Sheets []sheet `json:"sheets"`
}

type sheet struct {
	Name string                         `json:"name"`
	Data [][]map[string]json.RawMessage `json:"data"`
}

type category struct {
	name        string
	subcategory string
}

// ParseFile reads a FoodKeeper JSON export and returns its products sorted by
// ID, along with warnings for malformed cells that were tolerated (parsed as
// absent). Embeddings are not populated.
func ParseFile(path string) ([]*models.FoodKeeperProduct, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	doc := &document{}
	if err := json.Unmarshal(data, doc); err != nil {
		return nil, nil, err
	}

	var categorySheet, productSheet *sheet
	for i := range doc.Sheets {
		switch doc.Sheets[i].Name {
		case "Category":
			categorySheet = &doc.Sheets[i]
		case "Product":
			productSheet = &doc.Sheets[i]
		}
	}
	if categorySheet == nil || productSheet == nil {
		return nil, nil, fmt.Errorf("missing Category or Product sheet")
	}

	var warnings []string

	categories := make(map[int64]category, len(categorySheet.Data))
	for i, cells := range categorySheet.Data {
		row := mergeRow(cells)
		id, err := cellInt64(row, "ID")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipping category row %d: %v", i, err))
			continue
		}
		categories[id] = category{
			name:        cellString(row, "Category_Name"),
			subcategory: cellString(row, "Subcategory_Name"),
		}
	}

	products := make([]*models.FoodKeeperProduct, 0, len(productSheet.Data))
	for i, cells := range productSheet.Data {
		row := mergeRow(cells)
		id, err := cellInt64(row, "ID")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipping product row %d: %v", i, err))
			continue
		}

		p := &models.FoodKeeperProduct{
			ID:           id,
			Name:         cellString(row, "Name"),
			NameSubtitle: cellString(row, "Name_subtitle"),
			Keywords:     cellString(row, "Keywords"),
		}

		p.CategoryID, err = cellInt64(row, "Category_ID")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("product %d: %v", id, err))
		} else if cat, ok := categories[p.CategoryID]; ok {
			p.CategoryName = cat.name
			p.SubcategoryName = cat.subcategory
		} else {
			warnings = append(warnings, fmt.Sprintf("product %d: unknown category ID %d", id, p.CategoryID))
		}

		// The source data's tips column casing is inconsistent (Pantry_tips
		// vs Freeze_Tips), so the exact key is listed per group; an empty
		// tips key means the group has no tips column.
		groups := []struct {
			prefix  string
			tipsKey string
			dst     *models.FoodKeeperDuration
		}{
			{"Pantry", "Pantry_tips", &p.Pantry},
			{"DOP_Pantry", "DOP_Pantry_tips", &p.DOPPantry},
			{"Pantry_After_Opening", "", &p.PantryAfterOpening},
			{"Refrigerate", "Refrigerate_tips", &p.Refrigerate},
			{"DOP_Refrigerate", "DOP_Refrigerate_tips", &p.DOPRefrigerate},
			{"Refrigerate_After_Opening", "", &p.RefrigerateAfterOpening},
			{"Refrigerate_After_Thawing", "", &p.RefrigerateAfterThawing},
			{"Freeze", "Freeze_Tips", &p.Freeze},
			{"DOP_Freeze", "DOP_Freeze_Tips", &p.DOPFreeze},
		}
		for _, g := range groups {
			if v, err := cellFloatPtr(row, g.prefix+"_Min"); err != nil {
				warnings = append(warnings, fmt.Sprintf("product %d: %v", id, err))
			} else {
				g.dst.Min = v
			}
			if v, err := cellFloatPtr(row, g.prefix+"_Max"); err != nil {
				warnings = append(warnings, fmt.Sprintf("product %d: %v", id, err))
			} else {
				g.dst.Max = v
			}
			g.dst.Metric = cellString(row, g.prefix+"_Metric")
			if g.tipsKey != "" {
				g.dst.Tips = cellString(row, g.tipsKey)
			}
		}

		products = append(products, p)
	}

	sort.Slice(products, func(a, b int) bool {
		return products[a].ID < products[b].ID
	})

	return products, warnings, nil
}

func mergeRow(cells []map[string]json.RawMessage) map[string]json.RawMessage {
	row := make(map[string]json.RawMessage, len(cells))
	for _, cell := range cells {
		maps.Copy(row, cell)
	}
	return row
}

// cellString returns the cell's string value, or "" when the cell is missing,
// null, or not a string.
func cellString(row map[string]json.RawMessage, key string) string {
	raw, ok := row[key]
	if !ok {
		return ""
	}
	var s *string
	if err := json.Unmarshal(raw, &s); err != nil || s == nil {
		return ""
	}
	return *s
}

// cellFloatPtr returns nil for missing/null cells and an error for cells that
// hold a non-numeric value (a handful of Min cells contain free text).
func cellFloatPtr(row map[string]json.RawMessage, key string) (*float64, error) {
	raw, ok := row[key]
	if !ok {
		return nil, nil
	}
	var f *float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("%s: non-numeric value %s", key, raw)
	}
	return f, nil
}

// cellInt64 extracts a required integer cell (stored as a JSON float).
func cellInt64(row map[string]json.RawMessage, key string) (int64, error) {
	raw, ok := row[key]
	if !ok {
		return 0, fmt.Errorf("missing %s", key)
	}
	var f *float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0, fmt.Errorf("%s: non-numeric value %s", key, raw)
	}
	if f == nil {
		return 0, fmt.Errorf("%s is null", key)
	}
	return int64(*f), nil
}
