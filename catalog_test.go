package postgres

import (
	"context"
	"encoding/json"
	"github.com/ChristianDenniss/go-data-model/catalog"
	"os"
	"testing"
)

// Only run against a disposable database; the test owns the catalog read tables.
func TestCatalogImportIntegration(t *testing.T) {
	url := os.Getenv("CATALOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set CATALOG_TEST_DATABASE_URL to a disposable database")
	}
	ctx := context.Background()
	db, err := Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewCatalogRepository(db)
	amount := int64(1530)
	b := catalog.Bundle{Version: 1, Providers: map[string]catalog.Snapshot{"Uber Eats": {City: "Fredericton", Region: "NB", RetrievedAt: "2026-09-26T12:00:00Z", Stores: []catalog.Store{{ID: "ss_ubereats_0a5ba1d349096949787465cf", Name: "Taco Boyz", URL: "https://www.ubereats.com/ca/store/taco-boyz/gnecSBVbV9ygUE84RWhwVA", ObservedAt: "2026-09-26T12:00:00Z", Items: []catalog.MenuItem{{Name: "Large Nachos", Amount: &amount, Currency: "CAD"}}}}}}}
	raw, _ := json.Marshal(b)
	if err = repo.Import(ctx, b, raw); err != nil {
		t.Fatal(err)
	}
	if err = repo.Import(ctx, b, raw); err != nil {
		t.Fatal(err)
	}
	var count int
	db.sql.QueryRowContext(ctx, `SELECT count(*) FROM catalog_imports`).Scan(&count)
	if count != 1 {
		t.Fatal("replay created duplicate history", count)
	}
	got, err := repo.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if *got.Providers["Uber Eats"].Stores[0].Items[0].Amount != 1530 {
		t.Fatal("round trip lost price")
	}
	snap := b.Providers["Uber Eats"]
	snap.Stores[0].Items[0].Currency = "USD"
	b.Providers["Uber Eats"] = snap
	bad, _ := json.Marshal(b)
	if repo.Import(ctx, b, bad) == nil {
		t.Fatal("database accepted invalid currency")
	}
	got, err = repo.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Providers["Uber Eats"].Stores[0].Items[0].Currency != "CAD" {
		t.Fatal("failed import changed current data")
	}
	snap.Stores[0].Items[0].Currency = "CAD"
	snap.Stores[0].ObservedAt = "2020-01-01T00:00:00Z"
	b.Providers["Uber Eats"] = snap
	older, _ := json.Marshal(b)
	if repo.Import(ctx, b, older) == nil {
		t.Fatal("older import rolled back prices")
	}
	db.sql.QueryRowContext(ctx, `SELECT count(*) FROM catalog_imports`).Scan(&count)
	if count != 1 {
		t.Fatal("failed import left history rows")
	}
}
