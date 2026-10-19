package analytics

type Sale struct {
	Product          string
	Category         string
	Quantity         int
	UnitPriceKopecks int64
	Cancelled        bool
}

type Report struct {
	RevenueKopecks    int64
	UnitsSold         int
	RevenueByCategory map[string]int64
	TopProduct        string
	InvalidSales      int
}

func AnalyzeSales(sales []Sale) Report {
	report := Report{}
	report.RevenueByCategory = make(map[string]int64)

	if len(sales) == 0 {
		// чтобы возвращалась именно пустая map, а не nil (при return Report{})
		return report
	}

	quantities := make(map[string]int)

	for _, s := range sales {
		if s.Cancelled {
			continue
		}

		if s.Product == "" || s.Category == "" || s.Quantity <= 0 || s.UnitPriceKopecks <= 0 {
			report.InvalidSales++
			continue
		}

		saleRevenue := int64(s.Quantity) * s.UnitPriceKopecks
		report.RevenueKopecks += saleRevenue
		report.UnitsSold += s.Quantity
		report.RevenueByCategory[s.Category] += saleRevenue

		quantities[s.Product] += s.Quantity
	}

	var topProduct string
	var bestQuantity int

	for p, q := range quantities {
		if q > bestQuantity {
			topProduct = p
			bestQuantity = q
		} else if q == bestQuantity {
			topProduct = min(topProduct, p)
		}
	}

	report.TopProduct = topProduct

	return report
}
