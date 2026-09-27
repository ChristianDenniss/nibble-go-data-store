package postgres

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/ChristianDenniss/go-data-model/restaurant/entity"
)

// openTestDB connects to NIBBLE_TEST_DATABASE_URL (migrations run on open) or skips the test.
// Point it at a throwaway database: tests write and delete rows.
func openTestDB(t *testing.T) *DB {
	t.Helper()
	url := os.Getenv("NIBBLE_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("NIBBLE_TEST_DATABASE_URL not set")
	}
	db, err := Open(context.Background(), url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func testRestaurant(t *testing.T, db *DB, id string) entity.Restaurant {
	t.Helper()
	t.Cleanup(func() {
		_, _ = db.sql.Exec(`DELETE FROM restaurants WHERE id = $1`, id)
	})
	return entity.Restaurant{ID: id, Name: "Hours Test " + id}
}

func TestRestaurantHoursRoundTrip(t *testing.T) {
	db := openTestDB(t)
	repo := NewRestaurantRepository(db)
	ctx := context.Background()

	in := testRestaurant(t, db, "rest_test_hours_roundtrip")
	in.Hours = []entity.Hours{
		{Service: entity.HoursStore, DayOfWeek: 6, Opens: "17:00", Closes: "02:00"},
		{Service: entity.HoursStore, DayOfWeek: 6, Opens: "11:00", Closes: "14:30"},
		{Service: entity.HoursDelivery, DayOfWeek: 0, Opens: "12:00", Closes: "21:00"},
	}
	if err := repo.Upsert(ctx, in); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := repo.GetByID(ctx, in.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	want := []entity.Hours{
		{Service: entity.HoursDelivery, DayOfWeek: 0, Opens: "12:00", Closes: "21:00"},
		{Service: entity.HoursStore, DayOfWeek: 6, Opens: "11:00", Closes: "14:30"},
		{Service: entity.HoursStore, DayOfWeek: 6, Opens: "17:00", Closes: "02:00"},
	}
	if !reflect.DeepEqual(got.Hours, want) {
		t.Fatalf("hours = %+v, want %+v", got.Hours, want)
	}

	catalog, err := (&StorefrontRepository{db: db}).listRestaurants(ctx)
	if err != nil {
		t.Fatalf("list restaurants: %v", err)
	}
	for _, r := range catalog {
		if r.ID == in.ID {
			if !reflect.DeepEqual(r.Hours, want) {
				t.Fatalf("storefront hours = %+v, want %+v", r.Hours, want)
			}
			return
		}
	}
	t.Fatalf("restaurant %s missing from storefront list", in.ID)
}

func TestRestaurantUpsertHoursReplaceKeepAndClear(t *testing.T) {
	db := openTestDB(t)
	repo := NewRestaurantRepository(db)
	ctx := context.Background()

	in := testRestaurant(t, db, "rest_test_hours_replace")
	in.Hours = []entity.Hours{{Service: entity.HoursStore, DayOfWeek: 1, Opens: "09:00", Closes: "17:00"}}
	if err := repo.Upsert(ctx, in); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	replaced := []entity.Hours{{Service: entity.HoursDelivery, DayOfWeek: 2, Opens: "10:00", Closes: "16:00"}}
	in.Hours = replaced
	if err := repo.Upsert(ctx, in); err != nil {
		t.Fatalf("replace: %v", err)
	}
	assertHours(t, repo, in.ID, replaced)

	in.Hours = nil
	if err := repo.Upsert(ctx, in); err != nil {
		t.Fatalf("upsert without hours: %v", err)
	}
	assertHours(t, repo, in.ID, replaced)

	in.Hours = []entity.Hours{}
	if err := repo.Upsert(ctx, in); err != nil {
		t.Fatalf("clear: %v", err)
	}
	assertHours(t, repo, in.ID, nil)
}

func TestRestaurantHoursRejectsInvalidRows(t *testing.T) {
	db := openTestDB(t)
	repo := NewRestaurantRepository(db)
	ctx := context.Background()

	cases := map[string]entity.Hours{
		"unknown service":  {Service: "pickup", DayOfWeek: 1, Opens: "09:00", Closes: "17:00"},
		"day out of range": {Service: entity.HoursStore, DayOfWeek: 7, Opens: "09:00", Closes: "17:00"},
		"hour 24":          {Service: entity.HoursStore, DayOfWeek: 1, Opens: "09:00", Closes: "24:00"},
		"unpadded hour":    {Service: entity.HoursStore, DayOfWeek: 1, Opens: "9:00", Closes: "17:00"},
	}
	for name, hours := range cases {
		t.Run(name, func(t *testing.T) {
			in := testRestaurant(t, db, "rest_test_hours_invalid")
			in.Hours = []entity.Hours{hours}
			if err := repo.Upsert(ctx, in); err == nil {
				t.Fatalf("upsert accepted %+v", hours)
			}
		})
	}
}

func TestRestaurantHoursCascadeOnDelete(t *testing.T) {
	db := openTestDB(t)
	repo := NewRestaurantRepository(db)
	ctx := context.Background()

	in := testRestaurant(t, db, "rest_test_hours_cascade")
	in.Hours = []entity.Hours{{Service: entity.HoursStore, DayOfWeek: 3, Opens: "08:00", Closes: "20:00"}}
	if err := repo.Upsert(ctx, in); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if _, err := db.sql.ExecContext(ctx, `DELETE FROM restaurants WHERE id = $1`, in.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	var left int
	if err := db.sql.QueryRowContext(ctx, `SELECT count(*) FROM restaurant_hours WHERE restaurant_id = $1`, in.ID).Scan(&left); err != nil {
		t.Fatal(err)
	}
	if left != 0 {
		t.Fatalf("%d hours rows left after deleting the restaurant", left)
	}
}

func assertHours(t *testing.T, repo *RestaurantRepository, id string, want []entity.Hours) {
	t.Helper()
	got, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !reflect.DeepEqual(got.Hours, want) {
		t.Fatalf("hours = %+v, want %+v", got.Hours, want)
	}
}
