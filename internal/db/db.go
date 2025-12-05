package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/pressly/goose/v3"
	"github.com/s-588/BOMViewer/cmd/config"
	db "github.com/s-588/BOMViewer/internal/db/generate"
	"github.com/s-588/BOMViewer/internal/models"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrNotFound       = errors.New("ничего не найдено")
	ErrInternal       = errors.New("внутреняя ошибка")
	ErrAlreadyExist   = errors.New("такой объект уже существует")
	ErrMustBeFilled   = errors.New("обязательные поля должны быть заполнены")
	ErrIncorrectValue = errors.New("введено некоректное значение")

	//go:embed sql/migrations/*.sql
	embededMigrations embed.FS
)

type Repository struct {
	queries *db.Queries
	db      *sql.DB
}

func NewRepository(ctx context.Context, cfg config.Config) (*Repository, error) {
	connStr := filepath.Join(cfg.BaseDirectory, cfg.DBCfg.DBName)
	conn, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("can't ping db: %w", err)
	}
	r := &Repository{
		queries: db.New(conn),
		db:      conn,
	}
	err = r.initDB()
	if err != nil {
		return nil, fmt.Errorf("can't initialize db: %w", err)
	}

	return r, conn.Ping()
}

func (r *Repository) initDB() error {
	goose.SetBaseFS(embededMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	if err := goose.Up(r.db, "sql/migrations"); err != nil {
		return err
	}
	return nil
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
	rows, err := r.queries.GetAllMaterials(ctx)
	if err != nil {
		return nil, parseError(err)
	}

	materialsMap := make(map[int64]*models.Material)

	for _, row := range rows {
		materialID := row.MaterialID

		if _, exists := materialsMap[materialID]; !exists {
			material := &models.Material{
				ID: materialID,
				Unit: models.Unit{
					ID:   row.UnitID,
					Name: row.Unit,
				},
				Description: row.Description.String,
				PrimaryName: row.MaterialName,
				Names:       []string{},
				Products:    []models.Product{},
			}

			if q, ok := row.Quantity.(string); ok {
				material.Quantity = q
			} else {
				material.Quantity = row.QuantityText.String
			}

			materialsMap[materialID] = material
		}

		// Add the current name to the material's names (avoid duplicates)
		currentName := row.MaterialName
		if currentName != "" {
			// Check if name already exists to avoid duplicates
			nameExists := false
			for _, existingName := range materialsMap[materialID].Names {
				if existingName == currentName {
					nameExists = true
					break
				}
			}
			if !nameExists {
				materialsMap[materialID].Names = append(materialsMap[materialID].Names, currentName)
			}
		}

		// Add product if it exists
		if row.ProductID.Valid {
			// Check if product already exists to avoid duplicates
			productExists := false
			for _, existingProduct := range materialsMap[materialID].Products {
				if existingProduct.ID == row.ProductID.Int64 {
					productExists = true
					break
				}
			}
			if !productExists {
				product := models.Product{
					ID:   row.ProductID.Int64,
					Name: row.ProductName.String,
				}
				materialsMap[materialID].Products = append(materialsMap[materialID].Products, product)
			}
		}
	}

	materials := make([]models.Material, 0, len(materialsMap))
	for _, material := range materialsMap {
		materials = append(materials, *material)
	}

	return materials, nil
}

func (r *Repository) GetAllMaterialsWithPrimaryNames(ctx context.Context) ([]models.Material, error) {
	rows, err := r.queries.GetAllMaterialsWithPrimaryNames(ctx)
	if err != nil {
		return nil, parseError(err)
	}

	materialsMap := make(map[int64]*models.Material)

	for _, row := range rows {
		materialID := row.MaterialID

		if _, exists := materialsMap[materialID]; !exists {
			material := &models.Material{
				ID: materialID,
				Unit: models.Unit{
					ID:   row.UnitID,
					Name: row.Unit,
				},
				Description: row.Description.String,
				PrimaryName: row.MaterialName,
				Names:       []string{},
				Products:    []models.Product{},
			}

			if q, ok := row.Quantity.(string); ok {
				material.Quantity = q
			} else {
				material.Quantity = row.QuantityText.String
			}

			materialsMap[materialID] = material
		}

		// Add product if it exists
		if row.ProductID.Valid {
			// Check if product already exists to avoid duplicates
			productExists := false
			for _, existingProduct := range materialsMap[materialID].Products {
				if existingProduct.ID == row.ProductID.Int64 {
					productExists = true
					break
				}
			}
			if !productExists {
				product := models.Product{
					ID:   row.ProductID.Int64,
					Name: row.ProductName.String,
				}
				materialsMap[materialID].Products = append(materialsMap[materialID].Products, product)
			}
		}
	}

	materials := make([]models.Material, 0, len(materialsMap))
	for _, material := range materialsMap {
		materials = append(materials, *material)
	}

	return materials, nil
}

func (r *Repository) InsertMaterial(ctx context.Context, material models.Material) (models.Material, error) {
	materialRow, err := r.queries.InsertMaterial(ctx, db.InsertMaterialParams{
		Unit: material.Unit.Name, // This might be the issue!
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
		_, err = r.queries.InsertMaterialName(ctx, db.InsertMaterialNameParams{
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
	_, err := r.queries.UpdateMaterialUnit(ctx, db.UpdateMaterialUnitParams{
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
		_, err := r.queries.InsertMaterialName(ctx, db.InsertMaterialNameParams{
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
			ID:       row.FileID,
			Name:     row.Name,
			Path:     row.Path,
			MimeType: row.MimeType,
			FileType: row.FileType.String,
		}
		files = append(files, file)
	}
	return files, nil
}

func (r *Repository) InsertMaterialFile(ctx context.Context, materialID, fileID int64) error {
	_, err := r.queries.InsertMaterialFile(ctx, db.InsertMaterialFileParams{
		MaterialID: sql.NullInt64{Int64: materialID, Valid: true},
		FileID:     sql.NullInt64{Int64: fileID, Valid: true},
	})
	return parseError(err)
}

func (r *Repository) InsertFile(ctx context.Context, file models.File) (int64, error) {
	fmt.Println(file)
	fileRow, err := r.queries.InsertFile(ctx, db.InsertFileParams{
		Name:     file.Name,
		Path:     file.Path,
		MimeType: file.MimeType,
		FileType: sql.NullString{String: file.FileType, Valid: true},
	})
	if err != nil {
		return 0, parseError(err)
	}
	return fileRow.FileID, nil
}

func (r *Repository) GetFileByID(ctx context.Context, id int64) (models.File, error) {
	fileRow, err := r.queries.GetFileByID(ctx, id)
	if err != nil {
		return models.File{}, parseError(err)
	}

	return models.File{
		ID:       fileRow.FileID,
		Name:     fileRow.Name,
		Path:     fileRow.Path,
		MimeType: fileRow.MimeType,
		FileType: fileRow.FileType.String,
	}, nil
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

	// Group by product ID and aggregate materials
	productsMap := make(map[int64]*models.Product)

	for _, row := range productRows {
		productID := row.ProductID

		// Create product if it doesn't exist
		if _, exists := productsMap[productID]; !exists {
			product := &models.Product{
				ID:          productID,
				Name:        row.Name,
				Description: row.Description.String,
				Materials:   []models.Material{},
			}
			productsMap[productID] = product
		}

		// Add material if it exists for this product
		if row.MaterialID.Valid {
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

			material := models.Material{
				ID: row.MaterialID.Int64,
				Unit: models.Unit{
					ID:   row.UnitID.Int64,
					Name: row.UnitName.String,
				},
				PrimaryName: row.MaterialName.String,
				Quantity:    quantity,
			}
			productsMap[productID].Materials = append(productsMap[productID].Materials, material)
		}
	}

	// Convert map to slice
	products := make([]models.Product, 0, len(productsMap))
	for _, product := range productsMap {
		products = append(products, *product)
	}

	return products, nil
}

func (r *Repository) InsertProduct(ctx context.Context, product models.Product) (int64, error) {
	row, err := r.queries.InsertProduct(ctx, db.InsertProductParams{
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
			ID:       row.FileID,
			Name:     row.Name,
			Path:     row.Path,
			MimeType: row.MimeType,
			FileType: row.FileType.String,
		})
	}
	return files, nil
}

func (r *Repository) InsertProductFile(ctx context.Context, fileID, productID int64) error {
	_, err := r.queries.InsertProductFile(ctx, db.InsertProductFileParams{
		FileID:    sql.NullInt64{Int64: fileID, Valid: true},
		ProductID: sql.NullInt64{Int64: productID, Valid: true},
	})
	return parseError(err)
}

func (r *Repository) DeleteProductFile(ctx context.Context, productID, fileID int64) error {
	err := r.queries.DeleteProductFile(ctx, db.DeleteProductFileParams{
		ProductID: sql.NullInt64{Int64: productID, Valid: true},
		FileID:    sql.NullInt64{Int64: fileID, Valid: true},
	})
	return parseError(err)
}

func (r *Repository) AddProductMaterial(ctx context.Context, productID, materialID int64, quantity string) error {
	req := db.AddProductMaterialParams{
		ProductID:  sql.NullInt64{Int64: productID, Valid: true},
		MaterialID: sql.NullInt64{Int64: materialID, Valid: true},
	}

	// Make quantity optional - only set if provided
	if quantity != "" {
		if quantityInt, err := strconv.ParseInt(quantity, 10, 64); err == nil {
			req.Quantity = quantityInt
		} else {
			req.QuantityText = sql.NullString{String: quantity, Valid: true}
		}
	}
	// If quantity is empty, both Quantity and QuantityText will be NULL/empty

	err := r.queries.AddProductMaterial(ctx, req)
	return parseError(err)
}

func (r *Repository) DeleteProductMaterial(ctx context.Context, productID, materialID int64) error {
	err := r.queries.DeleteProductMaterial(ctx, db.DeleteProductMaterialParams{
		ProductID:  sql.NullInt64{Int64: productID, Valid: true},
		MaterialID: sql.NullInt64{Int64: materialID, Valid: true},
	})
	return parseError(err)
}

func (r *Repository) UpdateMaterialDescription(ctx context.Context, materialID int64, description string) error {
	_, err := r.queries.UpdateMaterialDescription(ctx, db.UpdateMaterialDescriptionParams{
		MaterialID:  materialID,
		Description: sql.NullString{String: description, Valid: true},
	})
	return parseError(err)
}

func parseError(err error) error {
	if err == nil {
		return nil
	}
	slog.Debug("parsing database error", "error type", reflect.TypeOf(err), "error", err)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var sqliteError *sqlite.Error
	if errors.As(err, &sqliteError) {
		switch sqliteError.Code() {
		case sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY:
			return ErrNotFound
		case sqlite3.SQLITE_CONSTRAINT_UNIQUE:
			return ErrAlreadyExist
		case sqlite3.SQLITE_CONSTRAINT_NOTNULL:
			return ErrMustBeFilled
		case sqlite3.SQLITE_CONSTRAINT_CHECK:
			return ErrIncorrectValue
		}
	}
	return ErrInternal
}

func (r *Repository) UpdateProductName(ctx context.Context, id int64, name string) error {
	err := r.queries.UpdateProductName(ctx, db.UpdateProductNameParams{
		ProductID: id,
		Name:      name,
	})
	return parseError(err)
}

func (r *Repository) UpdateProductDescription(ctx context.Context, d int64, description string) error {
	err := r.queries.UpdateProductDescription(ctx, db.UpdateProductDescriptionParams{
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
	result, err := r.queries.SearchAll(ctx, db.SearchAllParams{
		Text:  q,
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
	result, err := r.queries.SearchMaterials(ctx, db.SearchMaterialsParams{
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
				ID:   item.UnitID,
				Name: item.Unit,
			},
			Description: item.Text,
			Quantity:    quantity,
		})
	}

	return materials, nil
}

func (r *Repository) SearchProducts(ctx context.Context, q string, limit int64) ([]models.Product, error) {
	result, err := r.queries.SearchProducts(ctx, db.SearchProductsParams{
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

func (r *Repository) GetMaterialProfilePicture(ctx context.Context, id int64) (models.File, error) {
	row, err := r.queries.GetMaterialProfilePicture(ctx, sql.NullInt64{Int64: id, Valid: true})
	return models.File{
		ID:       row.FileID,
		Name:     row.Name,
		Path:     row.Path,
		MimeType: row.MimeType,
		FileType: row.FileType.String,
	}, parseError(err)
}

func (r *Repository) GetProductProfilePicture(ctx context.Context, id int64) (models.File, error) {
	row, err := r.queries.GetProductProfilePicture(ctx, sql.NullInt64{Int64: id, Valid: true})
	return models.File{
		ID:       row.FileID,
		Name:     row.Name,
		Path:     row.Path,
		MimeType: row.MimeType,
		FileType: row.FileType.String,
	}, parseError(err)
}

func (r *Repository) DeleteMaterialFile(ctx context.Context, materialID, fileID int64) error {
	return parseError(r.queries.DeleteMaterialFile(ctx, db.DeleteMaterialFileParams{
		FileID:     sql.NullInt64{Int64: fileID, Valid: true},
		MaterialID: sql.NullInt64{Int64: materialID, Valid: true},
	}))
}

func (r *Repository) SetMaterialProfilePicture(ctx context.Context, materialID, fileID int64) error {
	err := r.queries.SetFileToProfilePicture(ctx, fileID)
	if err != nil {
		return parseError(err)
	}
	if err != nil {
		return parseError(err)
	}
	return nil
}

// In your db.go - fix the SetProductProfilePicture method
func (r *Repository) SetProductProfilePicture(ctx context.Context, productID, fileID int64) error {
	// First, check if the file is already associated with the product
	files, err := r.queries.GetProductFiles(ctx, sql.NullInt64{Int64: productID, Valid: true})
	if err != nil {
		return parseError(err)
	}

	// Check if file is already associated
	fileAlreadyAssociated := false
	for _, file := range files {
		if file.FileID == fileID {
			fileAlreadyAssociated = true
			break
		}
	}

	// Only insert if not already associated
	if !fileAlreadyAssociated {
		_, err = r.queries.InsertProductFile(ctx, db.InsertProductFileParams{
			ProductID: sql.NullInt64{Int64: productID, Valid: true},
			FileID:    sql.NullInt64{Int64: fileID, Valid: true},
		})
		if err != nil {
			return parseError(err)
		}
	}

	// Now set as profile picture
	err = r.queries.SetFileToProfilePicture(ctx, fileID)
	return parseError(err)
}
func (r *Repository) UnsetMaterialProfilePicture(ctx context.Context, materialID int64) error {
	return parseError(r.queries.UnsetMaterialProfilePicture(ctx, sql.NullInt64{Int64: materialID, Valid: true}))

}
func (r *Repository) UnsetProductProfilePicture(ctx context.Context, productID int64) error {
	return parseError(r.queries.UnsetProductProfilePicture(ctx, sql.NullInt64{Int64: productID, Valid: true}))
}
func (r *Repository) GetAllMaterialImages(ctx context.Context, materialID int64) ([]models.File, error) {
	imgRows, err := r.queries.GetAllMaterialImages(ctx, sql.NullInt64{Int64: materialID, Valid: true})
	if err != nil {
		return nil, parseError(err)
	}

	files := make([]models.File, 0, len(imgRows))
	for _, file := range imgRows {
		files = append(files, models.File{
			ID:       file.FileID,
			Name:     file.Name,
			Path:     file.Path,
			MimeType: file.MimeType,
			FileType: file.FileType.String,
		})
	}
	return files, nil
}
func (r *Repository) GetAllProductImages(ctx context.Context, productID int64) ([]models.File, error) {
	imgRows, err := r.queries.GetAllProductImages(ctx, sql.NullInt64{Int64: productID, Valid: true})
	if err != nil {
		return nil, parseError(err)
	}

	files := make([]models.File, 0, len(imgRows))
	for _, file := range imgRows {
		files = append(files, models.File{
			ID:       file.FileID,
			Name:     file.Name,
			Path:     file.Path,
			MimeType: file.MimeType,
			FileType: file.FileType.String,
		})
	}
	return files, nil
}

func (r *Repository) UpdateMaterialProducts(ctx context.Context, materialID int64, productIDs []models.Product) error {
	err := r.queries.DeleteAllMaterialProducts(ctx, sql.NullInt64{Int64: materialID, Valid: true})
	if err != nil {
		return parseError(err)
	}

	for _, product := range productIDs {
		args := db.AddProductMaterialParams{
			ProductID:  sql.NullInt64{Int64: product.ID, Valid: true},
			MaterialID: sql.NullInt64{Int64: materialID, Valid: true},
		}
		if q, err := strconv.ParseInt(product.Quantity, 10, 64); err == nil {
			args.Quantity = q
		} else {
			args.QuantityText = sql.NullString{String: product.Quantity, Valid: true}
		}
		err := r.queries.AddProductMaterial(ctx, args)
		if err != nil {
			return parseError(err)
		}
	}
	return nil
}

// UpdateProduct updates a product's basic information
func (r *Repository) UpdateProduct(ctx context.Context, id int64, name, description string) error {
	err := r.queries.UpdateProductName(ctx, db.UpdateProductNameParams{
		ProductID: id,
		Name:      name,
	})
	if err != nil {
		return parseError(err)
	}

	err = r.queries.UpdateProductDescription(ctx, db.UpdateProductDescriptionParams{
		ProductID:   id,
		Description: sql.NullString{String: description, Valid: description != ""},
	})
	return parseError(err)
}

// UpdateProductMaterials updates all materials for a product
func (r *Repository) UpdateProductMaterials(ctx context.Context, productID int64, materials []models.Material) error {
	// Delete all existing material associations
	err := r.queries.DeleteAllProductMaterials(ctx, sql.NullInt64{Int64: productID, Valid: true})
	if err != nil {
		return parseError(err)
	}

	// Add new material associations
	for _, material := range materials {
		args := db.AddProductMaterialParams{
			ProductID:  sql.NullInt64{Int64: productID, Valid: true},
			MaterialID: sql.NullInt64{Int64: material.ID, Valid: true},
		}

		// Handle quantity - similar to how materials handle it
		if material.Quantity != "" {
			if q, err := strconv.ParseInt(material.Quantity, 10, 64); err == nil {
				args.Quantity = q
			} else {
				args.QuantityText = sql.NullString{String: material.Quantity, Valid: true}
			}
		}

		err := r.queries.AddProductMaterial(ctx, args)
		if err != nil {
			return parseError(err)
		}
	}
	return nil
}
