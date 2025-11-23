package helpers

import (
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/s-588/BOMViewer/internal/models"
)

// FilterArgs contains all possible filter parameters for materials
type MaterialFilterArgs struct {
	PrimaryOnly  bool
	ProductIDs   []int64
	UnitIDs      []int64
	MinQuantity  *float64
	MaxQuantity  *float64
	QuantityUnit string // to handle different units for quantity range
}

// ProductFilterArgs contains all possible filter parameters for products
type ProductFilterArgs struct {
	MaterialIDs  []int64
	NameContains string
}

// SortConfig defines sorting parameters
type SortConfig struct {
	Field string // "name", "unit", "quantity", "id"
	Order string // "asc", "desc"
}

// FilterMaterials applies all filters to materials list
func FilterMaterials(materials []models.Material, filter MaterialFilterArgs) []models.Material {
	if len(materials) == 0 {
		return materials
	}

	filtered := make([]models.Material, 0, len(materials))

	for _, material := range materials {
		if !passesMaterialFilters(material, filter) {
			continue
		}
		filtered = append(filtered, material)
	}

	return filtered
}

// passesMaterialFilters checks if a material passes all active filters
func passesMaterialFilters(material models.Material, filter MaterialFilterArgs) bool {
	// Primary name filter
	if filter.PrimaryOnly && material.PrimaryName == "" {
		slog.Debug("material is not passes the primary only filter", "material", material, "PrimaryOnly", filter.PrimaryOnly)
		return false
	}

	// Product filter - material must be used in specified products
	if len(filter.ProductIDs) > 0 {
		if !isMaterialUsedInProducts(material, filter.ProductIDs) {
			slog.Debug("material is not passes the product filter", "material", material, "ProductIDs", filter.ProductIDs)
			return false
		}
	}

	// Unit filter
	if len(filter.UnitIDs) > 0 {
		if !contains(filter.UnitIDs, material.Unit.ID) {
			slog.Debug("material is not passes the unit filter", "material", material, "UnitIDs", filter.UnitIDs)
			return false
		}
	}

	// Quantity range filter
	if filter.MinQuantity != nil || filter.MaxQuantity != nil {
		if !passesQuantityFilter(material.Quantity, filter) {
			slog.Debug("material is not passes the quantity filter", "material", material, "MinQuantity", filter.MinQuantity, "MaxQuantity", filter.MaxQuantity)
			return false
		}
	}

	return true
}

// isMaterialUsedInProducts checks if material is used in any of the specified products
func isMaterialUsedInProducts(material models.Material, productIDs []int64) bool {
	for _, product := range material.Products {
		if contains(productIDs, product.ID) {
			return true
		}
	}
	return false
}

// passesQuantityFilter checks if material quantity falls within specified range
func passesQuantityFilter(quantityStr string, filter MaterialFilterArgs) bool {
	if quantityStr == "" {
		return false
	}

	// Parse quantity - handle both numeric and text quantities
	quantity, err := parseQuantity(quantityStr)
	if err != nil {
		return false // Skip materials with unparseable quantities
	}

	if filter.MinQuantity != nil && quantity < *filter.MinQuantity {
		return false
	}

	if filter.MaxQuantity != nil && quantity > *filter.MaxQuantity {
		return false
	}

	return true
}

// parseQuantity attempts to parse quantity string to float64
func parseQuantity(quantityStr string) (float64, error) {
	// Clean the string - remove spaces, commas as decimal separators, etc.
	cleaned := strings.ReplaceAll(quantityStr, ",", ".")
	cleaned = strings.TrimSpace(cleaned)

	// Try parsing as float
	return strconv.ParseFloat(cleaned, 64)
}

// SortMaterials sorts materials based on sort configuration
func SortMaterials(materials []models.Material, config SortConfig) {
	if len(materials) <= 1 {
		return
	}

	sort.Slice(materials, func(i, j int) bool {
		switch config.Field {
		case "name":
			return compareStrings(materials[i].PrimaryName, materials[j].PrimaryName, config.Order)
		case "unit":
			return compareStrings(materials[i].Unit.Name, materials[j].Unit.Name, config.Order)
		case "quantity":
			return compareQuantities(materials[i].Quantity, materials[j].Quantity, config.Order)
		case "id":
			return compareInts(int(materials[i].ID), int(materials[j].ID), config.Order)
		default:
			return compareStrings(materials[i].PrimaryName, materials[j].PrimaryName, config.Order)
		}
	})
}
