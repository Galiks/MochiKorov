package game

import (
	"encoding/json"
	"fmt"
	"os"
)

type CardStoreData struct {
	Establishments []Card `json:"establishments"`
	Landmarks      []Card `json:"landmarks"`
}

func SaveCardDefinitions(path string) error {
	data := CardStoreData{
		Establishments: AllEstablishments(),
		Landmarks:      AllLandmarks(),
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("card store marshal: %w", err)
	}
	return os.WriteFile(path, raw, 0644)
}

func LoadCardDefinitions(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("card store read: %w", err)
	}

	var data CardStoreData
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("card store unmarshal: %w", err)
	}

	establishmentRegistry = make(map[string]Card)
	landmarkRegistry = make(map[string]Card)

	registerDefaultCards()

	for _, c := range data.Establishments {
		if c.ID == "" {
			continue
		}
		if _, exists := establishmentRegistry[c.ID]; exists {
			if c.DefaultStock == 0 {
				c.DefaultStock = 6
			}
			establishmentRegistry[c.ID] = c
		}
	}

	for _, c := range data.Landmarks {
		if c.ID == "" {
			continue
		}
		if _, exists := landmarkRegistry[c.ID]; exists {
			landmarkRegistry[c.ID] = c
		}
	}

	return nil
}

func MergeCardDefinitions(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("card store read: %w", err)
	}

	var data CardStoreData
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("card store unmarshal: %w", err)
	}

	for _, c := range data.Establishments {
		if c.ID == "" {
			continue
		}
		RegisterEstablishment(c)
	}

	for _, c := range data.Landmarks {
		if c.ID == "" {
			continue
		}
		RegisterLandmark(c)
	}

	return nil
}
