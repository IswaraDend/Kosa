package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// importRowError reports one failed row so the whole file can be rejected
// with actionable feedback (all-or-nothing — see parse* functions below).
type importRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type itemStockImportRow struct {
	itemID      uint
	warehouseID uint
	quantity    float64
	unitCost    float64
}

type productStockImportRow struct {
	productID   uint
	warehouseID uint
	quantity    float64
	unitCost    float64 // 0 means "not provided" — quantity-only correction
}

func openImportFile(c *gin.Context) (*excelize.File, string, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, "", fmt.Errorf("file excel wajib diupload (field 'file')")
	}
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".xlsx") {
		return nil, "", fmt.Errorf("file harus berformat .xlsx")
	}

	f, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuka file")
	}
	defer f.Close()

	xf, err := excelize.OpenReader(f)
	if err != nil {
		return nil, "", fmt.Errorf("file excel tidak valid atau rusak")
	}

	return xf, fileHeader.Filename, nil
}

func readSheetRows(xf *excelize.File) ([][]string, error) {
	sheets := xf.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("file excel tidak memiliki sheet")
	}
	rows, err := xf.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("gagal membaca isi file excel")
	}
	return rows, nil
}

func rowCell(row []string, idx int) string {
	if idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func parseCellFloat(v string) (float64, error) {
	if v == "" {
		return 0, nil
	}
	return strconv.ParseFloat(v, 64)
}

// parseItemStockImport validates the "SKU | Gudang | Qty | Harga Beli per Unit"
// columns row by row. All-or-nothing: any row error means nothing gets written.
func parseItemStockImport(rows [][]string, projectID uint) ([]itemStockImportRow, []importRowError) {
	if len(rows) < 2 {
		return nil, []importRowError{{Row: 1, Message: "File tidak memiliki baris data"}}
	}

	var parsed []itemStockImportRow
	var errs []importRowError

	for i, row := range rows[1:] {
		rowNum := i + 2
		sku := rowCell(row, 0)
		if sku == "" {
			continue // skip fully blank trailing rows
		}
		warehouseCode := rowCell(row, 1)
		qtyStr := rowCell(row, 2)
		costStr := rowCell(row, 3)

		var item models.Item
		if err := database.DB.Where("project_id = ? AND sku = ?", projectID, sku).First(&item).Error; err != nil {
			errs = append(errs, importRowError{Row: rowNum, Message: fmt.Sprintf("SKU '%s' tidak ditemukan di project ini", sku)})
			continue
		}

		var warehouse models.Warehouse
		if err := database.DB.Where("project_id = ? AND code = ?", projectID, warehouseCode).First(&warehouse).Error; err != nil {
			errs = append(errs, importRowError{Row: rowNum, Message: fmt.Sprintf("Gudang '%s' tidak ditemukan di project ini", warehouseCode)})
			continue
		}

		qty, err := parseCellFloat(qtyStr)
		if err != nil || qty <= 0 {
			errs = append(errs, importRowError{Row: rowNum, Message: "Qty harus berupa angka lebih dari 0"})
			continue
		}

		cost, err := parseCellFloat(costStr)
		if err != nil || cost < 0 {
			errs = append(errs, importRowError{Row: rowNum, Message: "Harga beli per unit harus berupa angka"})
			continue
		}
		if cost == 0 {
			// Falls back to the item's existing average cost — only valid if
			// it already has one (an item receiving stock for the first time
			// has no cost to fall back to).
			if item.AverageCost == 0 {
				errs = append(errs, importRowError{Row: rowNum, Message: fmt.Sprintf("Harga beli per unit wajib diisi untuk SKU '%s' (item ini belum pernah punya harga)", sku)})
				continue
			}
			cost = item.AverageCost
		}

		parsed = append(parsed, itemStockImportRow{itemID: item.ID, warehouseID: warehouse.ID, quantity: qty, unitCost: cost})
	}

	return parsed, errs
}

// parseProductStockImport validates "SKU | Gudang | Qty | HPP per Unit
// (opsional)". Unlike items, cost is genuinely optional here — a blank cost
// means "quantity-only correction", not an error.
func parseProductStockImport(rows [][]string, projectID uint) ([]productStockImportRow, []importRowError) {
	if len(rows) < 2 {
		return nil, []importRowError{{Row: 1, Message: "File tidak memiliki baris data"}}
	}

	var parsed []productStockImportRow
	var errs []importRowError

	for i, row := range rows[1:] {
		rowNum := i + 2
		sku := rowCell(row, 0)
		if sku == "" {
			continue
		}
		warehouseCode := rowCell(row, 1)
		qtyStr := rowCell(row, 2)
		costStr := rowCell(row, 3)

		var product models.Product
		if err := database.DB.Where("project_id = ? AND sku = ?", projectID, sku).First(&product).Error; err != nil {
			errs = append(errs, importRowError{Row: rowNum, Message: fmt.Sprintf("SKU '%s' tidak ditemukan di project ini", sku)})
			continue
		}

		var warehouse models.Warehouse
		if err := database.DB.Where("project_id = ? AND code = ?", projectID, warehouseCode).First(&warehouse).Error; err != nil {
			errs = append(errs, importRowError{Row: rowNum, Message: fmt.Sprintf("Gudang '%s' tidak ditemukan di project ini", warehouseCode)})
			continue
		}

		qty, err := parseCellFloat(qtyStr)
		if err != nil || qty <= 0 {
			errs = append(errs, importRowError{Row: rowNum, Message: "Qty harus berupa angka lebih dari 0"})
			continue
		}

		cost, err := parseCellFloat(costStr)
		if err != nil || cost < 0 {
			errs = append(errs, importRowError{Row: rowNum, Message: "HPP per unit harus berupa angka"})
			continue
		}

		parsed = append(parsed, productStockImportRow{productID: product.ID, warehouseID: warehouse.ID, quantity: qty, unitCost: cost})
	}

	return parsed, errs
}

// executeItemStockImport replays every valid row as a bulk "in" Transaction
// (grouped by warehouse) so Item.AverageCost updates through the exact same
// path as a manually entered transaction, with full audit-trail Transaction/
// TransactionItem rows. One DB transaction — any failure rolls back everything.
func executeItemStockImport(projectID, callerID uint, rows []itemStockImportRow, fileName string) (int, error) {
	groups := map[uint][]TransactionItemRequest{}
	for _, r := range rows {
		groups[r.warehouseID] = append(groups[r.warehouseID], TransactionItemRequest{
			ItemID:   r.itemID,
			Quantity: r.quantity,
			UnitCost: r.unitCost,
		})
	}

	note := fmt.Sprintf("Import Excel: %s", fileName)
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for warehouseID, items := range groups {
			if _, err := createTransactionTx(tx, projectID, callerID, models.TransactionIn, nil, &warehouseID, note, items); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// executeProductStockImport is a direct ProductStock upsert (opening
// balance / correction) — it does NOT go through createProductionCore since
// no BOM consumption happens here.
func executeProductStockImport(projectID uint, rows []productStockImportRow) (int, error) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, r := range rows {
			if r.unitCost > 0 {
				if err := applyProductCostLayer(tx, r.productID, r.quantity, r.unitCost); err != nil {
					return err
				}
			}
			if err := adjustProductStock(tx, projectID, r.warehouseID, r.productID, r.quantity); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func handleItemStockImport(c *gin.Context, projectID uint) {
	xf, fileName, err := openImportFile(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rows, err := readSheetRows(xf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsed, rowErrors := parseItemStockImport(rows, projectID)
	if len(rowErrors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": rowErrors})
		return
	}
	if len(parsed) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Tidak ada baris data untuk diimport"})
		return
	}

	imported, err := executeItemStockImport(projectID, currentUserID(c), parsed, fileName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported_rows": imported})
}

func handleProductStockImport(c *gin.Context, projectID uint) {
	xf, _, err := openImportFile(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rows, err := readSheetRows(xf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsed, rowErrors := parseProductStockImport(rows, projectID)
	if len(rowErrors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": rowErrors})
		return
	}
	if len(parsed) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Tidak ada baris data untuk diimport"})
		return
	}

	imported, err := executeProductStockImport(projectID, parsed)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported_rows": imported})
}

// ---- Super Admin routes (project_id trusted from the multipart form field) ----

func ImportItemsStock(c *gin.Context) {
	projectID, ok := parseFormUint(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id wajib diisi"})
		return
	}
	handleItemStockImport(c, projectID)
}

func ImportProductsStock(c *gin.Context) {
	projectID, ok := parseFormUint(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id wajib diisi"})
		return
	}
	handleProductStockImport(c, projectID)
}

// ---- Admin routes (project trusted from context, set by middleware) ----

func ImportItemsStockForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	handleItemStockImport(c, projectID)
}

func ImportProductsStockForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	handleProductStockImport(c, projectID)
}

func parseFormUint(c *gin.Context, key string) (uint, bool) {
	v, err := strconv.ParseUint(c.PostForm(key), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(v), true
}
