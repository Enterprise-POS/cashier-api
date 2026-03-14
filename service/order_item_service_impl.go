package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/xuri/excelize/v2"
)

type OrderItemServiceImpl struct {
	Repository        repository.OrderItemRepository
	ItemNameRegexRule *regexp.Regexp
}

func NewOrderItemServiceImpl(repository repository.OrderItemRepository) OrderItemService {
	return &OrderItemServiceImpl{
		Repository:        repository,
		ItemNameRegexRule: regexp.MustCompile(`^[\p{Han}\p{Hiragana}\p{Katakana}a-zA-Z][\p{Han}\p{Hiragana}\p{Katakana}a-zA-Z0-9' ]*$`),
	}
}

// Get implements OrderItemService.
func (service *OrderItemServiceImpl) Get(
	tenantId int,
	storeId int,
	limit int,
	page int,
	filters []*query.QueryFilter,
	dateFilter *query.DateFilter,
) ([]*model.OrderItem, int, error) {
	if tenantId <= 0 {
		return nil, 0, errors.New("Tenant id is Required !")
	}

	if limit < 1 {
		return nil, 0, fmt.Errorf("Limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}

	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}

	// Date filter validation
	if dateFilter != nil {
		// Check if both dates are provided and start is after end
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			if *dateFilter.StartDate > *dateFilter.EndDate {
				return nil, 0, fmt.Errorf("Start date (%d) cannot be after end date (%d)", *dateFilter.StartDate, *dateFilter.EndDate)
			}
		}

		// Check for negative timestamps (dates before 1970)
		if dateFilter.StartDate != nil && *dateFilter.StartDate < 0 {
			return nil, 0, fmt.Errorf("Invalid start date timestamp: %d", *dateFilter.StartDate)
		}
		if dateFilter.EndDate != nil && *dateFilter.EndDate < 0 {
			return nil, 0, fmt.Errorf("Invalid emd date timestamp: %d", *dateFilter.EndDate)
		}

		// Check for unreasonably far future dates (e.g., year 2100+)
		maxTimestamp := int64(4102444800) // 2100-01-01 00:00:00 UTC
		if dateFilter.StartDate != nil && *dateFilter.StartDate > maxTimestamp {
			return nil, 0, fmt.Errorf("Start date is too far in the future: %d", *dateFilter.StartDate)
		}
		if dateFilter.EndDate != nil && *dateFilter.EndDate > maxTimestamp {
			return nil, 0, fmt.Errorf("End date is too far in the future: %d", *dateFilter.EndDate)
		}

		// Check for user intentionally specify endDate bigger than startDate
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			if *dateFilter.StartDate > *dateFilter.EndDate {
				return nil, 0, fmt.Errorf("Start date (%d) cannot be after end date (%d)", *dateFilter.StartDate, *dateFilter.EndDate)
			}
		}
	}

	orderItems, count, err := service.Repository.Get(tenantId, storeId, limit, page-1, filters, dateFilter)
	if err != nil {
		return nil, 0, err
	}

	return orderItems, count, nil
}

// PlaceOrderItem implements OrderItemService.
func (service *OrderItemServiceImpl) PlaceOrderItem(*model.OrderItem) (*model.OrderItem, error) {
	panic("unimplemented")
}

// Transactions implements OrderItemService.
func (service *OrderItemServiceImpl) Transactions(params *repository.CreateTransactionParams) (int, error) {
	if params.TenantId <= 0 || params.StoreId <= 0 || params.UserId <= 0 {
		return 0, errors.New("Tenant id, Store id, User id is Required !")
	}

	if len(params.Items) == 0 {
		return 0, errors.New("At least one item is required")
	}

	if len(params.Items) > 1000 {
		return 0, errors.New("Too many items (max 1000)")
	}

	var (
		calculatedSubTotal    = 0                 // Sum before discount
		calculatedDiscount    = 0                 // Total discount
		calculatedTotal       = 0                 // Sum after discount
		calculatedQuantity    = 0                 // Total quantity
		priceConsistencyCheck = make(map[int]int) // item_id -> price
	)
	for _, item := range params.Items {
		// Check price consistency for same item
		if existingPrice, exists := priceConsistencyCheck[item.ItemId]; exists {
			if existingPrice != item.StorePriceSnapshot {
				return 0, fmt.Errorf("Price mismatch for item_id %d: expected %d, got %d",
					item.ItemId, existingPrice, item.StorePriceSnapshot)
			}
		} else {
			priceConsistencyCheck[item.ItemId] = item.StorePriceSnapshot
		}

		if item.Quantity < 1 {
			return 0, fmt.Errorf("Given quantity %d, from item_id: %d. Quantity should never be <= 0", item.Quantity, item.Id)
		}

		if !service.ItemNameRegexRule.MatchString(item.ItemNameSnapshot) {
			// This is the same regex with WarehouseService.CreateItem
			return 0, fmt.Errorf("Illegal input from item name snapshot: %s", item.ItemNameSnapshot)
		}

		// Calculate totals
		itemSubTotal := item.StorePriceSnapshot * item.Quantity
		itemDiscount := item.DiscountAmount * item.Quantity
		itemTotal := itemSubTotal - itemDiscount

		calculatedSubTotal += itemSubTotal
		calculatedDiscount += itemDiscount
		calculatedTotal += itemTotal
		calculatedQuantity += item.Quantity

		// Validate individual item total
		if item.TotalAmount != itemTotal {
			return 0, fmt.Errorf("Item %d total mismatch: expected %d, got %d",
				item.ItemId, itemTotal, item.TotalAmount)
		}
	}

	// Validate against provided totals
	if calculatedQuantity != params.TotalQuantity {
		return 0, fmt.Errorf("Total quantity mismatch: calculated %d, provided %d",
			calculatedQuantity, params.TotalQuantity)
	}

	if calculatedSubTotal != params.SubTotal {
		return 0, fmt.Errorf("Subtotal mismatch: calculated %d, provided %d",
			calculatedSubTotal, params.SubTotal)
	}

	if calculatedTotal != params.TotalAmount {
		return 0, fmt.Errorf("Total amount mismatch: calculated %d, provided %d",
			calculatedTotal, params.TotalAmount)
	}

	if calculatedDiscount != params.DiscountAmount {
		return 0, fmt.Errorf("Discount amount mismatch: calculated %d, provided %d",
			calculatedDiscount, params.DiscountAmount)
	}

	// Validate payment (if you track cash given)
	// Remove this if PurchasedPrice is just another name for TotalAmount
	if params.PurchasedPrice < params.TotalAmount {
		return 0, fmt.Errorf("Insufficient payment: need %d, got %d",
			params.TotalAmount, params.PurchasedPrice)
	}

	orderId, err := service.Repository.Transactions(params)
	if err != nil {
		return 0, fmt.Errorf("Failed to create transaction: %w", err)
	}

	return orderId, nil
}

// FindById implements OrderItemService.
func (service *OrderItemServiceImpl) FindById(orderItemid int, tenantId int) (*model.OrderItem, []*model.PurchasedItem, error) {
	if tenantId <= 0 || orderItemid <= 0 {
		return nil, nil, errors.New("Tenant id or Order item id Required !")
	}

	orderItem, purchasedItemList, err := service.Repository.FindById(orderItemid, tenantId)
	if err != nil {
		return nil, nil, err
	}

	return orderItem, purchasedItemList, nil
}

// GetSalesReport implements OrderItemService.
func (service *OrderItemServiceImpl) GetSalesReport(tenantId int, storeId int, dateFilter *query.DateFilter) (*repository.SalesReport, error) {
	if tenantId <= 0 {
		return nil, errors.New("Tenant id is Required !")
	}
	// storeId = 0 is allowed, this allow to get report from all store
	if storeId < 0 {
		return nil, fmt.Errorf("Given store id value is not allowed. storeId: %d", storeId)
	}

	// dateFilter is allowed to nil
	// Date filter validation
	if dateFilter != nil {
		// Check if both dates are provided and start is after end
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			if *dateFilter.StartDate > *dateFilter.EndDate {
				return nil, fmt.Errorf("Start date (%d) cannot be after end date (%d)", *dateFilter.StartDate, *dateFilter.EndDate)
			}
		}

		// Check for negative timestamps (dates before 1970)
		if dateFilter.StartDate != nil && *dateFilter.StartDate < 0 {
			return nil, fmt.Errorf("Invalid start date timestamp: %d", *dateFilter.StartDate)
		}
		if dateFilter.EndDate != nil && *dateFilter.EndDate < 0 {
			return nil, fmt.Errorf("Invalid emd date timestamp: %d", *dateFilter.EndDate)
		}

		// Check for unreasonably far future dates (e.g., year 2100+)
		maxTimestamp := int64(4102444800) // 2100-01-01 00:00:00 UTC
		if dateFilter.StartDate != nil && *dateFilter.StartDate > maxTimestamp {
			return nil, fmt.Errorf("Start date is too far in the future: %d", *dateFilter.StartDate)
		}
		if dateFilter.EndDate != nil && *dateFilter.EndDate > maxTimestamp {
			return nil, fmt.Errorf("End date is too far in the future: %d", *dateFilter.EndDate)
		}

		// Check for user intentionally specify endDate bigger than startDate
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			if *dateFilter.StartDate > *dateFilter.EndDate {
				return nil, fmt.Errorf("Start date (%d) cannot be after end date (%d)", *dateFilter.StartDate, *dateFilter.EndDate)
			}
		}
	}

	salesReport, err := service.Repository.GetSalesReport(tenantId, storeId, dateFilter)
	if err != nil {
		return nil, err
	}

	return salesReport, nil
}

// ExportProfitExcel implements OrderItemService.
func (service *OrderItemServiceImpl) ExportProfitExcel(tenantId int, storeId int, dateFilter *query.DateFilter) ([]byte, error) {
	if tenantId <= 0 {
		return nil, errors.New("tenant id is required")
	}
	if storeId < 0 {
		return nil, fmt.Errorf("invalid store id: %d", storeId)
	}

	tenantName, storeName, err := service.Repository.GetTenantAndStoreName(tenantId, storeId)
	if err != nil {
		return nil, err
	}

	rows, err := service.Repository.GetProfitReport(tenantId, storeId, dateFilter)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	// ── Sheet 1: Per-Item Profit ──────────────────────────────────────────────
	itemSheet := "Profit Per Item"
	f.SetSheetName("Sheet1", itemSheet)

	// Header style: bold + center
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Currency style (IDR, no decimal)
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 3, // #,##0
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	// Profit column style with green/red conditional logic skipped for simplicity — use currency
	profitStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 3,
		Font:   &excelize.Font{Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	marginStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 2, // 0.00
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	headers := []string{"#", "Item Name", "Qty Sold", "Revenue (Rp)", "COGS (Rp)", "Discount (Rp)", "Profit (Rp)", "Margin (%)"}
	colWidths := []float64{5, 35, 12, 18, 18, 18, 18, 14}

	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(itemSheet, cell, h)
		f.SetCellStyle(itemSheet, cell, cell, headerStyle)
		f.SetColWidth(itemSheet, col, col, colWidths[i])
	}
	f.SetRowHeight(itemSheet, 1, 20)

	var (
		grandRevenue  int
		grandCogs     int
		grandDiscount int
		grandProfit   int
		grandQty      int
	)

	for i, row := range rows {
		excelRow := i + 2
		margin := 0.0
		if row.TotalRevenue > 0 {
			margin = float64(row.TotalProfit) / float64(row.TotalRevenue) * 100
		}

		cells := []interface{}{i + 1, row.ItemName, row.TotalQuantity, row.TotalRevenue, row.TotalCogs, row.TotalDiscount, row.TotalProfit, margin}
		for j, val := range cells {
			col, _ := excelize.ColumnNumberToName(j + 1)
			cell := fmt.Sprintf("%s%d", col, excelRow)
			f.SetCellValue(itemSheet, cell, val)
			switch j {
			case 3, 4, 5: // Revenue, COGS, Discount
				f.SetCellStyle(itemSheet, cell, cell, currencyStyle)
			case 6: // Profit
				f.SetCellStyle(itemSheet, cell, cell, profitStyle)
			case 7: // Margin
				f.SetCellStyle(itemSheet, cell, cell, marginStyle)
			}
		}

		grandRevenue += row.TotalRevenue
		grandCogs += row.TotalCogs
		grandDiscount += row.TotalDiscount
		grandProfit += row.TotalProfit
		grandQty += row.TotalQuantity
	}

	// Total row
	totalRow := len(rows) + 2
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true},
		NumFmt: 3,
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"D9E1F2"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
	})

	totalCells := map[string]interface{}{
		"A": "TOTAL", "B": "", "C": grandQty,
		"D": grandRevenue, "E": grandCogs, "F": grandDiscount, "G": grandProfit,
	}
	for col, val := range totalCells {
		cell := fmt.Sprintf("%s%d", col, totalRow)
		f.SetCellValue(itemSheet, cell, val)
		f.SetCellStyle(itemSheet, cell, cell, totalStyle)
	}
	grandMargin := 0.0
	if grandRevenue > 0 {
		grandMargin = float64(grandProfit) / float64(grandRevenue) * 100
	}
	marginTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true},
		NumFmt: 2,
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"D9E1F2"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
	})
	f.SetCellValue(itemSheet, fmt.Sprintf("H%d", totalRow), grandMargin)
	f.SetCellStyle(itemSheet, fmt.Sprintf("H%d", totalRow), fmt.Sprintf("H%d", totalRow), marginTotalStyle)

	// ── Sheet 2: Summary ──────────────────────────────────────────────────────
	summarySheet := "Summary"
	f.NewSheet(summarySheet)

	labelStyle, _ := f.NewStyle(&excelize.Style{
		Font:  &excelize.Font{Bold: true},
		Fill:  excelize.Fill{Type: "pattern", Color: []string{"EBF0FA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})
	valueStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt:    3,
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	valueTextStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "1F3864"},
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})

	f.SetColWidth(summarySheet, "A", "A", 28)
	f.SetColWidth(summarySheet, "B", "B", 22)

	f.SetCellValue(summarySheet, "A1", "Profit Report")
	f.SetCellStyle(summarySheet, "A1", "A1", titleStyle)
	f.MergeCell(summarySheet, "A1", "B1")

	generatedAt := time.Now().Format("02 Jan 2006 15:04:05")
	summaryRows := [][]interface{}{
		{"Generated At", generatedAt},
		{"Tenant", tenantName},
		{"Store", storeName},
		{},
		{"Total Items Sold", grandQty},
		{"Total Revenue (Rp)", grandRevenue},
		{"Total COGS (Rp)", grandCogs},
		{"Total Discount (Rp)", grandDiscount},
		{"Gross Profit (Rp)", grandProfit},
		{"Profit Margin (%)", fmt.Sprintf("%.2f%%", grandMargin)},
	}

	for i, sr := range summaryRows {
		row := i + 3
		if len(sr) == 0 {
			continue
		}
		labelCell := fmt.Sprintf("A%d", row)
		valueCell := fmt.Sprintf("B%d", row)
		f.SetCellValue(summarySheet, labelCell, sr[0])
		f.SetCellStyle(summarySheet, labelCell, labelCell, labelStyle)
		f.SetCellValue(summarySheet, valueCell, sr[1])
		if _, ok := sr[1].(int); ok {
			f.SetCellStyle(summarySheet, valueCell, valueCell, valueStyle)
		} else {
			f.SetCellStyle(summarySheet, valueCell, valueCell, valueTextStyle)
		}
	}

	// ── Chart 1: Clustered Column on "Profit Per Item" ───────────────────────
	if len(rows) > 0 {
		lastDataRow := len(rows) + 1 // row 1 = header, rows 2..N = data
		cat := fmt.Sprintf("'%s'!$B$2:$B$%d", itemSheet, lastDataRow)
		f.AddChart(itemSheet, "J1", &excelize.Chart{
			Type: excelize.Col,
			Series: []excelize.ChartSeries{
				{
					Name:       fmt.Sprintf("'%s'!$D$1", itemSheet),
					Categories: cat,
					Values:     fmt.Sprintf("'%s'!$D$2:$D$%d", itemSheet, lastDataRow),
					Fill:       excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1}, // blue
				},
				{
					Name:       fmt.Sprintf("'%s'!$E$1", itemSheet),
					Categories: cat,
					Values:     fmt.Sprintf("'%s'!$E$2:$E$%d", itemSheet, lastDataRow),
					Fill:       excelize.Fill{Type: "pattern", Color: []string{"ED7D31"}, Pattern: 1}, // orange
				},
				{
					Name:       fmt.Sprintf("'%s'!$G$1", itemSheet),
					Categories: cat,
					Values:     fmt.Sprintf("'%s'!$G$2:$G$%d", itemSheet, lastDataRow),
					Fill:       excelize.Fill{Type: "pattern", Color: []string{"70AD47"}, Pattern: 1}, // green
				},
			},
			Title:     []excelize.RichTextRun{{Text: "Revenue vs COGS vs Profit per Item"}},
			Legend:    excelize.ChartLegend{Position: "bottom"},
			Dimension: excelize.ChartDimension{Width: 640, Height: 360},
		})
	}

	// ── Chart 2: Pie on "Summary" (revenue breakdown) ────────────────────────
	// Write a small helper table in columns D/E for the pie series.
	f.SetColWidth(summarySheet, "D", "D", 18)
	f.SetColWidth(summarySheet, "E", "E", 18)

	pieHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellValue(summarySheet, "D2", "Component")
	f.SetCellValue(summarySheet, "E2", "Amount (Rp)")
	f.SetCellStyle(summarySheet, "D2", "E2", pieHeaderStyle)

	pieData := [][]interface{}{
		{"COGS", grandCogs},
		{"Discount", grandDiscount},
		{"Net Profit", grandProfit},
	}
	pieCurrencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 3, // #,##0
	})
	for i, pd := range pieData {
		row := i + 3
		f.SetCellValue(summarySheet, fmt.Sprintf("D%d", row), pd[0])
		f.SetCellValue(summarySheet, fmt.Sprintf("E%d", row), pd[1])
		f.SetCellStyle(summarySheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), pieCurrencyStyle)
	}

	f.AddChart(summarySheet, "D7", &excelize.Chart{
		Type: excelize.Bar,
		Series: []excelize.ChartSeries{
			{
				Name:   fmt.Sprintf("'%s'!$D$3", summarySheet),
				Values: fmt.Sprintf("'%s'!$E$3", summarySheet),
				Fill:   excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1}, // blue  - COGS
			},
			{
				Name:   fmt.Sprintf("'%s'!$D$4", summarySheet),
				Values: fmt.Sprintf("'%s'!$E$4", summarySheet),
				Fill:   excelize.Fill{Type: "pattern", Color: []string{"ED7D31"}, Pattern: 1}, // orange - Discount
			},
			{
				Name:   fmt.Sprintf("'%s'!$D$5", summarySheet),
				Values: fmt.Sprintf("'%s'!$E$5", summarySheet),
				Fill:   excelize.Fill{Type: "pattern", Color: []string{"70AD47"}, Pattern: 1}, // green  - Net Profit
			},
		},
		Title:     []excelize.RichTextRun{{Text: "Revenue Breakdown"}},
		Legend:    excelize.ChartLegend{Position: "bottom"},
		Dimension: excelize.ChartDimension{Width: 400, Height: 300},
	})

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel buffer: %w", err)
	}

	return buf.Bytes(), nil
}
