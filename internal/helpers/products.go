package helpers

import (
	"sort"
	"strings"

	"github.com/s-588/BOMViewer/internal/models"
)

// FilterProducts applies all filters to products list
func FilterProducts(products []models.Product, filter ProductFilterArgs) []models.Product {
	if len(products) == 0 {
		return products
	}

	filtered := make([]models.Product, 0, len(products))

	for _, product := range products {
		if !passesProductFilters(product, filter) {
			continue
		}
		filtered = append(filtered, product)
	}

	return filtered
}

// passesProductFilters checks if a product passes all active filters
func passesProductFilters(product models.Product, filter ProductFilterArgs) bool {
	// Name contains filter
	if filter.NameContains != "" {
		if !strings.Contains(strings.ToLower(product.Name), strings.ToLower(filter.NameContains)) {
			return false
		}
	}

	// Material filter - product must use specified materials
	if len(filter.MaterialIDs) > 0 {
		if !isProductUsingMaterials(product, filter.MaterialIDs) {
			return false
		}
	}

	return true
}

// isProductUsingMaterials checks if product uses any of the specified materials
func isProductUsingMaterials(product models.Product, materialIDs []int64) bool {
	for _, material := range product.Materials {
		if contains(materialIDs, material.ID) {
			return true
		}
	}
	return false
}

// SortProducts sorts products based on sort configuration
func SortProducts(products []models.Product, config SortConfig) {
	if len(products) <= 1 {
		return
	}

	sort.Slice(products, func(i, j int) bool {
		switch config.Field {
		case "name":
			return compareStrings(products[i].Name, products[j].Name, config.Order)
		case "id":
			return compareInts(int(products[i].ID), int(products[j].ID), config.Order)
		default:
			return compareStrings(products[i].Name, products[j].Name, config.Order)
		}
	})
}
