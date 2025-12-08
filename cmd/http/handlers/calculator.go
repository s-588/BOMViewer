package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/s-588/BOMViewer/internal/helpers"
	"github.com/s-588/BOMViewer/internal/models"
	"github.com/s-588/BOMViewer/web/templates"
)

type CalculatorFormData struct {
	ProductID           int64
	DesiredQuantity     int
	RemainingQuantities map[int64]string // materialID -> remaining quantity
}

type CalculationResult struct {
	MaterialID       int64
	MaterialName     string
	RequiredPerUnit  string
	RequiredTotal    string
	Remaining        string
	CanProduce       int
	AdditionalNeeded string
	Unit             string
	ProductID        int64
	IsCalculable     bool
}

func (h *Handler) CalculatorPageHandler(w http.ResponseWriter, r *http.Request) {
	products, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		helpers.SetAndLogError(w, http.StatusInternalServerError, 
			"ошибка получения списка изделии", 
			"can't get products for calculator page", "error", err)
		return
	}

	templates.CalculatorPage(products, []models.CalculationResult{}, []models.Material{}, 1).Render(r.Context(), w)
}

func (h *Handler) CalculatorProductMaterialsHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.SetAndLogError(w, http.StatusBadRequest, "Неверный идентификатор изделия", "invalid product ID in calculator product materials handler", err)
		return
	}

	// Get desired quantity from query params
	var desiredQuantity int64 = 1
	if dq := r.URL.Query().Get("desired_quantity"); dq != "" {
		if dqInt, err := strconv.ParseInt(dq, 10, 64); err == nil && dqInt > 0 {
			desiredQuantity = dqInt
		}
	}

	product, err := h.db.GetProductByID(r.Context(), productID)
	if err != nil {
		helpers.SetAndLogError(w, http.StatusInternalServerError, "ошибка получения материалов изделия", "error getting product materials in calculator product materials handler", err)
		return
	}

	// Split materials into calculable and non-calculable
	calculableMaterials, nonCalculableMaterials := h.splitMaterialsByCalculability(product.Materials)

	if len(calculableMaterials) == 0 && len(nonCalculableMaterials) == 0 {
		slog.Debug("no materials for calculator", "calculable materials", calculableMaterials, "non calculable materials", nonCalculableMaterials, "materials", product.Materials)
		templates.CalculatorResults([]models.CalculationResult{}, []models.Material{}, productID, desiredQuantity).Render(r.Context(), w)
		return
	}

	// Parse remaining quantities from query parameters (this is the fix)
	remainingQuantities := make(map[int64]float64)
	for _, material := range calculableMaterials {
		// Try to get remaining quantity from query params
		remainingStr := r.URL.Query().Get("remaining_" + strconv.FormatInt(material.ID, 10))
		if remainingStr != "" {
			if remaining, err := strconv.ParseFloat(remainingStr, 64); err == nil && remaining >= 0 {
				remainingQuantities[material.ID] = remaining
				continue
			}
		}
		// Default to 0 if not provided or invalid
		remainingQuantities[material.ID] = 0
	}

	// Process calculable materials
	var calculableResults []models.CalculationResult
	if len(calculableMaterials) > 0 {
		calculableResults = h.calculateMaterialRequirements(calculableMaterials, desiredQuantity, remainingQuantities, productID)
	}

	slog.Debug("results of calculator", "calculable results", calculableMaterials, "non calculable materials", nonCalculableMaterials, "calculable results", calculableResults)
	templates.CalculatorResults(calculableResults, nonCalculableMaterials, productID, desiredQuantity).Render(r.Context(), w)
}

func (h *Handler) CalculatorCalculateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		helpers.SetAndLogError(w, http.StatusInternalServerError, "ошибка обработки формы", "error parsing form in calculator calculate handler", "error", err)
		return
	}

	productID, err := strconv.ParseInt(r.FormValue("product_id"), 10, 64)
	if err != nil {
		helpers.SetAndLogError(w, http.StatusBadRequest, "Неверный идентификатор изделия", "invalid product ID in calculator calculate handler", "error", err)
		return
	}

	dq := r.FormValue("desired_quantity")
	if dq == "" {
		dq = "1"
	}
	desiredQuantity, err := strconv.ParseInt(dq, 10, 64)
	if err != nil || desiredQuantity <= 0 {
		helpers.SetAndLogError(w, http.StatusBadRequest, "Неверное желаемое количество", "invalid desired quantity in calculator calculate handler", "error", err)
		return
	}

	// Get product materials to know the required quantities
	materials, err := h.db.GetProductMaterials(r.Context(), productID)
	if err != nil {
		helpers.SetAndLogError(w, http.StatusInternalServerError, "ошибка получения материалов изделия", "error getting product materials in calculator calculate handler", "error", err)
		return
	}

	// Split materials into calculable and non-calculable
	calculableMaterials, nonCalculableMaterials := h.splitMaterialsByCalculability(materials)

	// Parse remaining quantities from form (only for calculable materials)
	remainingQuantities := make(map[int64]float64)
	for _, material := range calculableMaterials {
		remainingStr := r.FormValue("remaining_" + strconv.FormatInt(material.ID, 10))
		if remaining, err := strconv.ParseFloat(remainingStr, 64); err == nil && remaining >= 0 {
			remainingQuantities[material.ID] = remaining
		} else {
			remainingQuantities[material.ID] = 0
		}
	}

	// Calculate results only for calculable materials
	var calculableResults []models.CalculationResult
	if len(calculableMaterials) > 0 {
		calculableResults = h.calculateMaterialRequirements(calculableMaterials, desiredQuantity, remainingQuantities, productID)
	}

	templates.CalculatorResults(calculableResults, nonCalculableMaterials, productID, desiredQuantity).Render(r.Context(), w)
}

// splitMaterialsByCalculability separates materials into calculable (numeric quantities) and non-calculable (text quantities)
func (h *Handler) splitMaterialsByCalculability(materials []models.Material) (calculable []models.Material, nonCalculable []models.Material) {
	for _, material := range materials {
		if h.isQuantityCalculable(material.Quantity) {
			calculable = append(calculable, material)
		} else {
			nonCalculable = append(nonCalculable, material)
		}
	}
	return
}

// isQuantityCalculable checks if a quantity string can be parsed as a number
func (h *Handler) isQuantityCalculable(quantity string) bool {
	// Clean the quantity string - replace comma with dot for decimal parsing
	cleaned := strings.ReplaceAll(quantity, ",", ".")
	cleaned = strings.TrimSpace(cleaned)

	// Try to parse as float
	_, err := strconv.ParseFloat(cleaned, 64)
	return err == nil
}

func (h *Handler) calculateMaterialRequirements(materials []models.Material, desiredQuantity int64, remainingQuantities map[int64]float64, productID int64) []models.CalculationResult {
	results := make([]models.CalculationResult, len(materials))

	if len(materials) == 0 {
		return results
	}

	for i, material := range materials {
		requiredPerUnit, err := h.parseQuantity(material.Quantity)
		if err != nil || requiredPerUnit <= 0 {
			// If we can't parse the quantity, set canProduce to 0
			results[i] = models.CalculationResult{
				MaterialID:       material.ID,
				MaterialName:     material.PrimaryName,
				RequiredPerUnit:  material.Quantity,
				RequiredTotal:    "N/A",
				Remaining:        h.formatQuantity(remainingQuantities[material.ID]),
				CanProduce:       0,
				AdditionalNeeded: "N/A",
				Unit:             material.Unit.Name,
				ProductID:        productID,
				IsCalculable:     false,
			}
			continue
		}

		remaining := remainingQuantities[material.ID]
		totalRequired := requiredPerUnit * float64(desiredQuantity)
		additionalNeeded := totalRequired - remaining
		if additionalNeeded < 0 {
			additionalNeeded = 0
		}

		// Calculate how many products can be produced with this material alone
		canProduce := 0
		if requiredPerUnit > 0 {
			canProduce = int(remaining / requiredPerUnit)
		}

		results[i] = models.CalculationResult{
			MaterialID:       material.ID,
			MaterialName:     material.PrimaryName,
			RequiredPerUnit:  material.Quantity,
			RequiredTotal:    h.formatQuantity(totalRequired),
			Remaining:        h.formatQuantity(remaining),
			CanProduce:       canProduce,
			AdditionalNeeded: h.formatQuantity(additionalNeeded),
			Unit:             material.Unit.Name,
			ProductID:        productID,
			IsCalculable:     true,
		}
	}

	return results
}

// Helper function to parse quantity strings
func (h *Handler) parseQuantity(quantityStr string) (float64, error) {
	// Clean the quantity string - replace comma with dot for decimal parsing
	cleaned := strings.ReplaceAll(quantityStr, ",", ".")
	cleaned = strings.TrimSpace(cleaned)

	return strconv.ParseFloat(cleaned, 64)
}

// Helper function to format quantity for display
func (h *Handler) formatQuantity(quantity float64) string {
	if quantity == float64(int(quantity)) {
		return strconv.Itoa(int(quantity))
	}
	return strconv.FormatFloat(quantity, 'f', 2, 64)
}
