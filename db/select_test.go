package dbutil

import (
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"

	stringutil "github.com/monstercat/golib/string"

	. "github.com/monstercat/pgnull"
)

func TestGetColumnsByTag(t *testing.T) {

	type rptype struct {
		ArtistsTitle         string         `db:"artists_title"`
		Brand                NullString     `db:"brand"`
		BrandId              NullInt        `db:"brand_id"`
		CatalogId            NullString     `db:"catalog_id"`
		CopyrightPLine       string         `db:"copyright_p_line"`
		CoverFileId          NullString     `db:"cover_file_id"`
		Created              time.Time      `db:"created"`
		FeaturedArtistsTitle string         `db:"featured_artists_title"`
		GRid                 string         `db:"grid" select:"coalesce"`
		GenrePrimary         string         `db:"genre_primary"`
		GenreSecondary       string         `db:"genre_secondary"`
		Id                   string         `db:"id"`
		LabelId              string         `db:"label_id"`
		LabelTitle           string         `db:"label_title"`
		LabelLine            string         `db:"label_line"`
		PrereleaseDate       NullTime       `db:"prerelease_date"`
		Public               bool           `db:"public"`
		ReleaseDate          time.Time      `db:"release_date"`
		RepresentativeId     string         `db:"representative_id"`
		RepresentativeName   NullString     `db:"representative_name"` //From the view
		SpotifyId            NullString     `db:"spotify_id"`
		Tags                 pq.StringArray `db:"tags"`
		TagusId              int            `db:"tagus_id"`
		Title                string         `db:"title"`
		Version              string         `db:"version"`
		Type                 string         `db:"type"`
		UPC                  string         `db:"upc" select:"coalesce"`
	}
	var rp rptype
	var rps []rptype

	cols := GetColumnsByTag(&rp, "t1.")
	if len(cols) != 27 {
		t.Fatal("expecting 27 columns")
	}
	for _, v := range cols {
		if strings.Index(v, "t1.") == -1 {
			t.Errorf("Expecting prefix %s - column name: %s", "t1.", v)
		}
	}

	cols = GetColumnsByTag(&rps, "t1.")
	if len(cols) != 27 {
		t.Fatal("expecting 27 columns")
	}
	for _, v := range cols {
		if strings.Index(v, "t1.") == -1 {
			t.Errorf("Expecting prefix %s - column name: %s", "t1.", v)
		}
	}

	testCols := func(cols map[string]string) {
		for k, v := range cols {
			var c string
			switch k {
			case "UPC":
				c = "upc"
			case "GRid":
				c = "grid"
			default:
				c = stringutil.CamelToSnakeCase(c)
			}

			if strings.Index(v, c) == -1 {
				t.Errorf("Expecting field name %s - column name: %s", c, v)
			}
		}
	}

	testCols(GetColumnsByTag(&rp, ""))

	// Second time should get from cache!
	testCols(GetColumnsByTag(&rp, ""))
}