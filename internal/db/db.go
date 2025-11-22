package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	sqlite "github.com/s-588/BOMViewer/internal/db/generate"
	"github.com/s-588/BOMViewer/internal/models"

	sqlite3 "github.com/mattn/go-sqlite3"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInternal       = errors.New("internal error")
	ErrAlreadyExist   = errors.New("already exists")
	ErrMustBeFilled   = errors.New("required field must be filled")
	ErrIncorrectValue = errors.New("incorrect value provided")
)

type Repository struct {
	queries *sqlite.Queries
	db      *sql.DB
}

func NewRepository(ctx context.Context, connStr string) (*Repository, error) {
	conn, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}

	return &Repository{
		queries: sqlite.New(conn),
		db:      conn,
	}, err
}

// Close function close database connection and assumed to call it before app stop.
func (r *Repository) Close() error {
	return r.db.Close()
}

type MaterialFilterArgs struct {
	PrimaryOnly bool
	Units       []int64
	Products    []int64
}

func (r *Repository) GetAllMaterials(ctx context.Context) ([]models.Material, error) {
	// Fetch all base materials without filters (simple static query).
	rows, err := r.queries.GetAllMaterials(ctx)
	if err != nil {
		return nil, parseError(err)
	}
 // Group by material ID and aggregate products
    materialsMap := make(map[int64]*models.Material)
    
    for _, row := range rows {
        materialID := row.MaterialID
        
        // Create material if it doesn't exist
        if _, exists := materialsMap[materialID]; !exists {
            material := &models.Material{
                ID:   materialID,
                Unit: models.Unit{
                    ID:   row.UnitID,
                    Name: row.Unit,
                },
                Description: row.Description.String,
                PrimaryName: row.PrimaryName,
                Names:       []string{row.PrimaryName},
                Products:    []models.Product{},
            }
            
            if q, ok := row.Quantity.(string); ok {
                material.Quantity = q
            } else {
                material.Quantity = row.QuantityText.String
            }
            
            materialsMap[materialID] = material
        }
        
        // Add product if it exists for this material
        if row.ProductID.Valid {
            product := models.Product{
                ID:   row.ProductID.Int64,
                Name: row.ProductName.String,
            }
            materialsMap[materialID].Products = append(materialsMap[materialID].Products, product)
        }
    }
    
    // Convert map to slice
    materials := make([]models.Material, 0, len(materialsMap))
    for _, material := range materialsMap {
        materials = append(materials, *material)
    }
    
    return materials, nil
}

func (r *Repository) InsertMaterial(ctx context.Context, material models.Material) (models.Material, error) {
	materialRow, err := r.queries.InsertMaterial(ctx, sqlite.InsertMaterialParams{
		Unit: material.Unit.Name,
		Description: sql.NullString{
			String: material.Description,
			Valid:  true,
		},
	})
	if err != nil {
		return models.Material{}, parseError(err)
	}
	material.ID = materialRow.MaterialID

	for _, name := range material.Names {
		_, err = r.queries.InsertMaterialName(ctx, sqlite.InsertMaterialNameParams{
			MaterialID: materialRow.MaterialID,
			Name:       name,
			IsPrimary:  name == material.PrimaryName,
		})
		if err != nil {
			return models.Material{}, parseError(err)
		}
	}

	return material, nil
}

func (r *Repository) GetMaterialByID(ctx context.Context, id int64) (models.Material, error) {
	materialRow, err := r.queries.GetMaterialByID(ctx, id)
	if err != nil {
		return models.Material{}, parseError(err)
	}

	nameRows, err := r.queries.GetMaterialNames(ctx, id)
	if err != nil {
		return models.Material{}, parseError(err)
	}
	var primaryName string
	names := make([]string, 0, len(nameRows))
	for _, name := range nameRows {
		names = append(names, name.Name)
		if name.IsPrimary {
			primaryName = name.Name
		}
	}

	productsRow, err := r.queries.GetMaterialProducts(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return models.Material{}, parseError(err)
	}
	products := make([]models.Product, 0, len(productsRow))
	for _, row := range productsRow {
		var quantity string
		switch qt := row.Quantity.(type) {
		case int64:
			quantity = fmt.Sprintf("%d", qt)
		case float64:
			quantity = fmt.Sprintf("%.3f", qt)
		case string:
			quantity = qt
		default:
			if row.QuantityText.Valid {
				quantity = row.QuantityText.String
			}
		}

		product := models.Product{
			ID:          row.ProductID,
			Name:        row.Name,
			Description: row.Description.String,
			Quantity:    quantity,
		}

		products = append(products, product)
	}

	return models.Material{
		ID:          materialRow.MaterialID,
		Names:       names,
		PrimaryName: primaryName,
		Description: materialRow.Description.String,
		Unit: models.Unit{
			ID:   materialRow.UnitID,
			Name: materialRow.Unit,
		},
		Products: products,
	}, nil
}

func (r *Repository) GetMaterialByName(ctx context.Context, name string) (models.Material, error) {
	materialRow, err := r.queries.GetMaterialByName(ctx, name)
	if err != nil {
		return models.Material{}, parseError(err)
	}

	nameRows, err := r.queries.GetMaterialNames(ctx, materialRow.MaterialID)
	if err != nil {
		return models.Material{}, parseError(err)
	}
	var primaryName string
	names := make([]string, 0, len(nameRows))
	for _, name := range nameRows {
		names = append(names, name.Name)
		if name.IsPrimary {
			primaryName = name.Name
		}
	}

	productsRow, err := r.queries.GetMaterialProducts(ctx, sql.NullInt64{Int64: materialRow.MaterialID, Valid: true})
	if err != nil {
		return models.Material{}, parseError(err)
	}
	products := make([]models.Product, 0, len(productsRow))
	for _, row := range productsRow {
		var quantity string
		switch qt := row.Quantity.(type) {
		case int64:
			quantity = fmt.Sprintf("%d", qt)
		case float64:
			quantity = fmt.Sprintf("%.3f", qt)
		case string:
			quantity = qt
		default:
			if row.QuantityText.Valid {
				quantity = row.QuantityText.String
			}
		}

		product := models.Product{
			ID:          row.ProductID,
			Name:        row.Name,
			Description: row.Description.String,
			Quantity:    quantity,
		}

		products = append(products, product)
	}

	return models.Material{
		ID: materialRow.MaterialID,
		Unit: models.Unit{
			ID:   materialRow.UnitID,
			Name: materialRow.Unit,
		},
		PrimaryName: primaryName,
		Names:       names,
		Description: materialRow.Description.String,
	}, nil
}

func (r *Repository) UpdateMaterialUnit(ctx context.Context, id int64, unit string) error {
	_, err := r.queries.UpdateMaterialUnit(ctx, sqlite.UpdateMaterialUnitParams{
		MaterialID: id,
		Unit:       unit,
	})
	return parseError(err)
}

func (r *Repository) UpdateMaterialNames(ctx context.Context, id int64, primaryName string, names []string) error {
	err := r.queries.DeleteAllMaterialNames(ctx, id)
	if err != nil {
		return parseError(err)
	}
	for _, name := range names {
		_, err := r.queries.InsertMaterialName(ctx, sqlite.InsertMaterialNameParams{
			MaterialID: id,
			Name:       name,
			IsPrimary:  name == primaryName,
		})
		if err != nil {
			return parseError(err)
		}
	}
	return nil
}

func (r *Repository) SetMaterialPrimaryName(ctx context.Context, name string) error {
	return parseError(r.queries.SetMaterialPrimaryName(ctx, name))
}

func (r *Repository) UnsetMaterialPrimaryName(ctx context.Context, id int64) error {
	return parseError(r.queries.UnsetMaterialPrimaryName(ctx, id))
}

func (r *Repository) GetMaterialProducts(ctx context.Context, id int64) ([]models.Product, error) {
	productRows, err := r.queries.GetMaterialProducts(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return nil, parseError(err)
	}

	products := make([]models.Product, 0, len(productRows))
	for _, row := range productRows {
		product := models.Product{
			ID:          row.ProductID,
			Name:        row.Name,
			Description: row.Description.String,
		}

		products = append(products, product)
	}

	return products, nil
}

func (r *Repository) GetMaterialFiles(ctx context.Context, id int64) ([]models.File, error) {
	fileRows, err := r.queries.GetMaterialFiles(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return nil, parseError(err)
	}

	files := make([]models.File, 0, len(fileRows))
	for _, row := range fileRows {
		file := models.File{
			ID:   row.FileID,
			Name: row.Name,
			Path: row.Path.String,
		}
		files = append(files, file)
	}
	return files, nil
}

func (r *Repository) InsertMaterialFile(ctx context.Context, materialID, fileID int64) error {
	_, err := r.queries.InsertMaterialFile(ctx, sqlite.InsertMaterialFileParams{
		MaterialID: sql.NullInt64{Int64: materialID, Valid: true},
		FileID:     sql.NullInt64{Int64: fileID, Valid: true},
	})
	return parseError(err)
}

func (r *Repository) InsertFile(ctx context.Context, file models.File) (int64, error) {
	fileRow, err := r.queries.InsertFile(ctx, sqlite.InsertFileParams{
		Name: file.Name,
		Path: sql.NullString{String: file.Path, Valid: true},
	})
	if err != nil {
		return 0, parseError(err)
	}
	return fileRow.FileID, nil
}

func (r *Repository) DeleteFile(ctx context.Context, id int64) error {
	return parseError(r.queries.DeleteFile(ctx, id))
}

func (r *Repository) DeleteMaterial(ctx context.Context, id int64) error {
	return parseError(r.queries.DeleteMaterial(ctx, id))
}

func (r *Repository) DeleteProduct(ctx context.Context, id int64) error {
	return parseError(r.queries.DeleteProduct(ctx, id))
}

func (r *Repository) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	productRows, err := r.queries.GetAllProducts(ctx)
	if err != nil {
		return nil, parseError(err)
	}

	products := make([]models.Product, 0, len(productRows))
	for _, row := range productRows {
		product := models.Product{
			ID:          row.ProductID,
			Name:        row.Name,
			Description: row.Description.String,
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *Repository) InsertProduct(ctx context.Context, product models.Product) (int64, error) {
	row, err := r.queries.InsertProduct(ctx, sqlite.InsertProductParams{
		Name:        product.Name,
		Description: sql.NullString{String: product.Description, Valid: true},
	})
	if err != nil {
		return 0, parseError(err)
	}
	return row.ProductID, nil
}

func (r *Repository) GetProductByID(ctx context.Context, id int64) (models.Product, error) {
	productRow, err := r.queries.GetProductByID(ctx, id)
	if err != nil {
		return models.Product{}, parseError(err)
	}

	materialsRow, err := r.queries.GetProductMaterials(ctx, sql.NullInt64{Int64: productRow.ProductID, Valid: true})
	if err != nil {
		return models.Product{}, parseError(err)
	}
	materials := make([]models.Material, 0, len(materialsRow))
	for _, row := range materialsRow {
		var quantity string
		switch qt := row.Quantity.(type) {
		case int64:
			quantity = fmt.Sprintf("%d", qt)
		case float64:
			quantity = fmt.Sprintf("%f", qt)
		case string:
			quantity = qt
		default:
			if row.QuantityText.Valid {
				quantity = row.QuantityText.String
			}
		}
		materials = append(materials, models.Material{
			ID: row.MaterialID,
			Unit: models.Unit{
				ID:   row.UnitID,
				Name: row.Unit,
			},
			Description: row.Description.String,
			PrimaryName: row.MaterialName,
			Quantity:    quantity,
		})
	}

	return models.Product{
		ID:          productRow.ProductID,
		Name:        productRow.Name,
		Description: productRow.Description.String,
		Materials:   materials,
	}, nil
}

func (r *Repository) GetProductMaterials(ctx context.Context, id int64) ([]models.Material, error) {
	materialRows, err := r.queries.GetProductMaterials(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return nil, parseError(err)
	}
	materials := make([]models.Material, 0, len(materialRows))
	for _, row := range materialRows {
		materials = append(materials, models.Material{
			ID: row.MaterialID,
			Unit: models.Unit{
				ID:   row.UnitID,
				Name: row.Unit,
			},
			Description: row.Description.String,
			PrimaryName: row.MaterialName,
		})
	}

	return materials, nil
}

func (r *Repository) GetProductFiles(ctx context.Context, id int64) ([]models.File, error) {
	fileRows, err := r.queries.GetProductFiles(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return nil, parseError(err)
	}
	files := make([]models.File, 0, len(fileRows))
	for _, row := range fileRows {
		files = append(files, models.File{
			ID:   row.FileID,
			Name: row.Name,
			Path: row.Path.String,
		})
	}
	return files, nil
}

func (r *Repository) InsertProductFile(ctx context.Context, fileID, productID int64) error {
	_, err := r.queries.InsertProductFile(ctx, sqlite.InsertProductFileParams{
		FileID:    sql.NullInt64{Int64: fileID, Valid: true},
		ProductID: sql.NullInt64{Int64: productID, Valid: true},
	})
	return parseError(err)
}

func (r *Repository) DeleteProductFile(ctx context.Context, id int64) error {
	err := r.queries.DeleteProductFile(ctx, id)
	return parseError(err)
}

func (r *Repository) AddProductMaterial(ctx context.Context, productID, materialID int64, quantity string) error {
	req := sqlite.AddProductMaterialParams{
		ProductID:  sql.NullInt64{Int64: productID, Valid: true},
		MaterialID: sql.NullInt64{Int64: materialID, Valid: true},
	}
	if quantity == "" {
		return errors.New("quantity is required")
	}
	if quantityInt, err := strconv.ParseInt(quantity, 10, 64); err != nil {
		req.QuantityText = sql.NullString{String: quantity, Valid: true}
	} else {
		req.Quantity = quantityInt
	}
	err := r.queries.AddProductMaterial(ctx, req)
	return parseError(err)
}

func (r *Repository) DeleteProductMaterial(ctx context.Context, productID, materialID int64) error {
	err := r.queries.DeleteProductMaterial(ctx, sqlite.DeleteProductMaterialParams{
		ProductID:  sql.NullInt64{Int64: productID, Valid: true},
		MaterialID: sql.NullInt64{Int64: materialID, Valid: true},
	})
	return parseError(err)
}

func (r *Repository) UpdateMaterialDescription(ctx context.Context, materialID int64, description string) error {
	_, err := r.queries.UpdateMaterialDescription(ctx, sqlite.UpdateMaterialDescriptionParams{
		MaterialID:  materialID,
		Description: sql.NullString{String: description, Valid: true},
	})
	return parseError(err)
}

func parseError(err error) error {
	if err != nil {
		slog.Debug("database error", "error", err)
	}
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows):
		return ErrNotFound
	case errors.Is(err, sqlite3.ErrConstraintForeignKey):
		return ErrNotFound
	case errors.Is(err, sqlite3.ErrConstraintUnique):
		return ErrAlreadyExist
	case errors.Is(err, sqlite3.ErrConstraintNotNull):
		return ErrMustBeFilled
	case errors.Is(err, sqlite3.ErrConstraintCheck):
		return ErrIncorrectValue
	default:
		return ErrInternal
	}
}

func (r *Repository) UpdateProductName(ctx context.Context, id int64, name string) error {
	err := r.queries.UpdateProductName(ctx, sqlite.UpdateProductNameParams{
		ProductID: id,
		Name:      name,
	})
	return parseError(err)
}

func (r *Repository) UpdateProductMaterials(ctx context.Context, id int64, materials []models.Material) error {
	err := r.queries.DeleteAllProductMaterials(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return parseError(err)
	}
	for _, m := range materials {
		err = r.queries.UpdateProductMaterial(ctx, sqlite.UpdateProductMaterialParams{
			ProductID:  sql.NullInt64{Int64: id, Valid: true},
			MaterialID: sql.NullInt64{Int64: m.ID, Valid: true},
			Quantity:   m.Quantity,
		})
		if err != nil {
			return parseError(err)
		}
	}
	return nil
}

func (r *Repository) UpdateProductDescription(ctx context.Context, d int64, description string) error {
	err := r.queries.UpdateProductDescription(ctx, sqlite.UpdateProductDescriptionParams{
		ProductID:   d,
		Description: sql.NullString{String: description, Valid: true},
	})
	return parseError(err)
}

func (r *Repository) GetAllUnits(ctx context.Context) ([]models.Unit, error) {
	unitRows, err := r.queries.GetAllUnits(ctx)
	if err != nil {
		return nil, parseError(err)
	}
	units := make([]models.Unit, 0, len(unitRows))
	for _, row := range unitRows {
		units = append(units, models.Unit{
			ID:   row.UnitID,
			Name: row.Unit,
		})
	}
	return units, nil
}

func (r *Repository) GetUnitByID(ctx context.Context, id int64) (models.Unit, error) {
	unit, err := r.queries.GetUnitByID(ctx, id)
	return models.Unit{
		ID:   unit.UnitID,
		Name: unit.Unit,
	}, parseError(err)
}

func (r *Repository) SearchAll(ctx context.Context, q string, limit int64) ([]models.Material, []models.Product, error) {
	result, err := r.queries.SearchAll(ctx, sqlite.SearchAllParams{
		Text: q,
		Limit: limit,
	})
	if err != nil {
		return nil, nil, parseError(err)
	}

	materials := make([]models.Material, 0)
	products := make([]models.Product, 0)
	for _, item := range result {
		id, err := strconv.ParseInt(item.RefID, 10, 64)
		if err != nil {
			return nil, nil, errors.New("can't parse id returned from search query")
		}

		name, ok := item.DisplayName.(string)
		if !ok {
			return nil, nil, errors.New("can't convert search result name to string")
		}

		if item.Type == "material" {
			materials = append(materials, models.Material{
				ID:          id,
				PrimaryName: name,
			})
		} else {
			products = append(products, models.Product{
				ID:   id,
				Name: name,
			})
		}
	}
	return materials, products, nil
}

func (r *Repository) SearchMaterials(ctx context.Context, q string, limit int64) ([]models.Material, error) {
	result, err := r.queries.SearchMaterials(ctx, sqlite.SearchMaterialsParams{
		Query: q,
		Limit: limit,
	})
	if err != nil {
		return nil, parseError(err)
	}

	materials := make([]models.Material, 0, len(result))
	for _, item := range result {
		id, err := strconv.ParseInt(item.RefID, 10, 64)
		if err != nil {
			return nil, parseError(err)
		}
		quantity := item.Quantity.(string)
		materials = append(materials, models.Material{
			ID:          id,
			PrimaryName: item.DisplayName.String,
			Unit: models.Unit{
				ID: item.UnitID,
				Name: item.Unit,
			},
			Description: item.Text,
			Quantity:    quantity,
		})
	}

	return materials, nil
}

func (r *Repository) SearchProducts(ctx context.Context, q string, limit int64) ([]models.Product, error) {
	result, err := r.queries.SearchProducts(ctx, sqlite.SearchProductsParams{
		Text:  q,
		Limit: limit,
	})
	if err != nil {
		return nil, parseError(err)
	}

	products := make([]models.Product, 0, len(result))
	for _, item := range result {
		id, err := strconv.ParseInt(item.RefID, 10, 64)
		if err != nil {
			return nil, parseError(err)
		}
		products = append(products, models.Product{
			ID:          id,
			Name:        item.DisplayName.String,
			Description: item.Text,
		})
	}

	return products, nil
}
