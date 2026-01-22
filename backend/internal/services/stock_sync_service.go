package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"psm-backend/internal/database"
	"time"
)

// TWSE API response structure
type TWSeStock struct {
	CompanyCode string `json:"公司代號"`
	CompanyName string `json:"公司名稱"`
	ShortName   string `json:"公司簡稱"`
	Industry    string `json:"產業別"`
}

// TPEx API response structure (for OTC stocks)
type TPExStock struct {
	Code string `json:"SecuritiesCompanyCode"`
	Name string `json:"CompanyName"`
}

type StockSyncService struct {
	db *database.DB
}

func NewStockSyncService(db *database.DB) *StockSyncService {
	return &StockSyncService{db: db}
}

// SyncFromTWSE fetches all listed stocks from TWSE Open API and syncs to database
func (s *StockSyncService) SyncFromTWSE(ctx context.Context) (int, error) {
	// TWSE Open API endpoint for listed companies
	url := "https://openapi.twse.com.tw/v1/opendata/t187ap03_L"

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch TWSE data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("TWSE API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var stocks []TWSeStock
	if err := json.Unmarshal(body, &stocks); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Industry mapping (產業別代碼 -> 產業名稱)
	industryMap := map[string]string{
		"01": "水泥",
		"02": "食品",
		"03": "塑膠",
		"04": "紡織",
		"05": "電機",
		"06": "電器電纜",
		"08": "玻璃陶瓷",
		"09": "造紙",
		"10": "鋼鐵",
		"11": "橡膠",
		"12": "汽車",
		"14": "建材營造",
		"15": "航運",
		"16": "觀光餐旅",
		"17": "金融保險",
		"18": "貿易百貨",
		"20": "其他",
		"21": "化學生技醫療",
		"22": "半導體",
		"23": "電腦週邊",
		"24": "光電",
		"25": "通信網路",
		"26": "電子零組件",
		"27": "電子通路",
		"28": "資訊服務",
		"29": "其他電子",
		"30": "文化創意",
		"31": "農業科技",
		"32": "電子商務",
		"33": "綠能環保",
		"34": "數位雲端",
		"35": "運動休閒",
		"36": "居家生活",
		"80": "管理股票",
		"9299": "存託憑證",
	}

	// Prepare batch insert
	count := 0
	for _, stock := range stocks {
		if stock.CompanyCode == "" || stock.ShortName == "" {
			continue
		}

		industry := industryMap[stock.Industry]
		if industry == "" {
			industry = "其他"
		}

		// Upsert stock (insert or update)
		query := `
			INSERT INTO taiwan_stocks (symbol, name, market, industry, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (symbol) 
			DO UPDATE SET 
				name = EXCLUDED.name,
				market = EXCLUDED.market,
				industry = EXCLUDED.industry,
				updated_at = EXCLUDED.updated_at
		`

		_, err := s.db.ExecContext(ctx, query,
			stock.CompanyCode,
			stock.ShortName,
			"TSE",
			industry,
			time.Now(),
			time.Now(),
		)

		if err != nil {
			// Log error but continue
			fmt.Printf("Failed to insert stock %s: %v\n", stock.CompanyCode, err)
			continue
		}
		count++
	}

	return count, nil
}

// SyncFromTPEx fetches OTC stocks from TPEx API
func (s *StockSyncService) SyncFromTPEx(ctx context.Context) (int, error) {
	// TPEx API endpoint
	url := "https://www.tpex.org.tw/openapi/v1/tpex_mainboard_peratio_analysis"

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch TPEx data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("TPEx API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var stocks []TPExStock
	if err := json.Unmarshal(body, &stocks); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	count := 0
	for _, stock := range stocks {
		if stock.Code == "" || stock.Name == "" {
			continue
		}

		// Skip ETF and bonds (usually start with 00 or 9)
		if len(stock.Code) >= 2 && (stock.Code[:2] == "00" || stock.Code[:1] == "9") {
			continue
		}

		query := `
			INSERT INTO taiwan_stocks (symbol, name, market, industry, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (symbol) 
			DO UPDATE SET 
				name = EXCLUDED.name,
				market = EXCLUDED.market,
				updated_at = EXCLUDED.updated_at
		`

		_, err := s.db.ExecContext(ctx, query,
			stock.Code,
			stock.Name,
			"OTC",
			"上櫃",
			time.Now(),
			time.Now(),
		)

		if err != nil {
			fmt.Printf("Failed to insert OTC stock %s: %v\n", stock.Code, err)
			continue
		}
		count++
	}

	return count, nil
}

// SyncAll syncs both TSE and OTC stocks
func (s *StockSyncService) SyncAll(ctx context.Context) (map[string]int, error) {
	result := make(map[string]int)

	tseCount, err := s.SyncFromTWSE(ctx)
	if err != nil {
		return nil, fmt.Errorf("TSE sync failed: %w", err)
	}
	result["tse"] = tseCount

	otcCount, err := s.SyncFromTPEx(ctx)
	if err != nil {
		return nil, fmt.Errorf("OTC sync failed: %w", err)
	}
	result["otc"] = otcCount

	result["total"] = tseCount + otcCount

	return result, nil
}
