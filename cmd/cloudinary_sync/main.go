// Command cloudinary_sync fetches images from Cloudinary folders via the Admin
// API and syncs them into the cards and card_images tables.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// folderMapping is the explicit folder -> card ID mapping supplied for the
// main wedding-card catalog. IDs here reflect the current state of the
// database, which has diverged from the original migration seed order.
type folderMapping struct {
	folder  string
	cardIDs []int64
}

var mappings = []folderMapping{
	{"forest-green-walima", []int64{17}},
	{"burgundy-arch-shalima", []int64{15, 12}},
	{"floral-vellum-transparency", []int64{13}},
	{"sage-and-gold-nikkah", []int64{5}},
	{"blush-mehndi-laser-cut", []int64{6}},
	{"sage-monogram-envelope", []int64{8}},
	{"royal-blue-silver-crest", []int64{10}},
	{"emerald-baraat-suite", []int64{11}},
	{"mughal-rose-wax-seal", []int64{16}},
	{"vibrant-dholki", []int64{4}},
	{"walima-passport-ticket-suite", []int64{53}},
	{"vellum-script-nikkah-suite", []int64{54}},
	{"terracotta-hummingbird-fold", []int64{55}},
	{"terracotta-botanical-rsvp-suite", []int64{56}},
	{"sage-peacock-garden-save-the-date", []int64{57}},
	{"sage-panther-garden-nikkah-suite", []int64{58}},
	{"ruby-gold-thread-suite", []int64{59}},
	{"monochrome-botanical-walima-suite", []int64{60}},
	{"mauve-sculpted-nikkah-suite", []int64{61}},
	{"lilac-ribbon-barat-suite", []int64{62}},
	{"ivory-rose-vellum-invite", []int64{63}},
	{"ivory-mughal-arch-suite", []int64{64}},
	{"ivory-arch-save-the-date", []int64{65}},
	{"grey-maroon-mughal-pattern", []int64{66}},
	{"grey-arch-botanical-suite", []int64{67}},
	{"golden-rose-foil-details-suite", []int64{68}},
	{"golden-ribbon-barat-mehndi-suite", []int64{69}},
	{"golden-mughal-garden-invite", []int64{70}},
	{"golden-jungle-chinoiserie-suite", []int64{71}},
	{"golden-elephant-chinoiserie-suite", []int64{72}},
	{"emerald-ribbon-monogram-suite", []int64{73}},
	{"emerald-mughal-portrait-suite", []int64{74}},
	{"emerald-mehendi-envelope", []int64{75}},
	{"emerald-acrylic-ribbon-box", []int64{76}},
	{"denim-blue-nikkah-suite", []int64{77}},
	{"crimson-pampas-wedding-suite", []int64{78}},
	{"crimson-embossed-floral-suite", []int64{79}},
	{"blush-wedding-passport", []int64{80}},
	{"blush-peony-vellum-suite", []int64{81}},
	{"mughal-rose-nikkahnama-frame", []int64{82}},
	{"silver-sage-nikkahnama-frame", []int64{83}},
	{"jewel-bloom-nikkahnama-frame", []int64{84}},
}

// bidBoxIDRange covers the bid-box cards whose Cloudinary folders (if any)
// are discovered dynamically by slugifying the card name.
const bidBoxIDStart, bidBoxIDEnd = 18, 23

type cloudinaryResource struct {
	PublicID string `json:"public_id"`
	Format   string `json:"format"`
	Version  int64  `json:"version"`
	Folder   string `json:"folder"`
}

type cloudinarySearchResponse struct {
	Resources  []cloudinaryResource `json:"resources"`
	NextCursor string               `json:"next_cursor"`
}

type cloudinaryClient struct {
	cloudName  string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

func newCloudinaryClient(cloudName, apiKey, apiSecret string) *cloudinaryClient {
	return &cloudinaryClient{
		cloudName:  cloudName,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// searchFolder returns all image resources under the given Cloudinary folder,
// sorted by public_id so results are stable across runs.
func (c *cloudinaryClient) searchFolder(folder string) ([]cloudinaryResource, error) {
	var all []cloudinaryResource
	cursor := ""

	for {
		body := map[string]any{
			"expression":  fmt.Sprintf(`folder="%s"`, folder),
			"max_results": 500,
			"sort_by":     []map[string]string{{"public_id": "asc"}},
		}
		if cursor != "" {
			body["next_cursor"] = cursor
		}

		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/resources/search", c.cloudName)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(c.apiKey, c.apiSecret)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			var errBody bytes.Buffer
			errBody.ReadFrom(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("cloudinary search for folder %q failed: %s: %s", folder, resp.Status, errBody.String())
		}

		var sr cloudinarySearchResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&sr)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}

		all = append(all, sr.Resources...)

		if sr.NextCursor == "" {
			break
		}
		cursor = sr.NextCursor
	}

	return all, nil
}

// buildImageURL constructs a Cloudinary delivery URL in the form
// https://res.cloudinary.com/{cloud}/image/upload/v{version}/{folder}/{filename}
func buildImageURL(cloudName string, r cloudinaryResource) string {
	filename := r.PublicID
	if idx := strings.LastIndex(filename, "/"); idx != -1 {
		filename = filename[idx+1:]
	}
	folder := r.Folder
	if folder == "" {
		if idx := strings.LastIndex(r.PublicID, "/"); idx != -1 {
			folder = r.PublicID[:idx]
		}
	}
	return fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/v%d/%s/%s.%s", cloudName, r.Version, folder, filename, r.Format)
}

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.ToLower(name)
	s = slugNonAlnum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type syncResult struct {
	folder string
	cardID int64
	count  int
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: could not load .env file: %v", err)
	}

	cloudName := strings.TrimSpace(os.Getenv("CLOUDINARY_CLOUD_NAME"))
	apiKey := strings.TrimSpace(os.Getenv("CLOUDINARY_API_KEY"))
	apiSecret := strings.TrimSpace(os.Getenv("CLOUDINARY_API_SECRET"))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		log.Fatal("CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, and CLOUDINARY_API_SECRET must be set")
	}
	if databaseURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	client := newCloudinaryClient(cloudName, apiKey, apiSecret)

	allMappings := append([]folderMapping{}, mappings...)
	allMappings = append(allMappings, discoverBidBoxMappings(ctx, pool, client)...)

	var results []syncResult

	for _, m := range allMappings {
		resources, err := client.searchFolder(m.folder)
		if err != nil {
			log.Printf("skipping folder %q: %v", m.folder, err)
			continue
		}
		if len(resources) == 0 {
			log.Printf("skipping folder %q: no images found", m.folder)
			continue
		}

		urls := make([]string, len(resources))
		for i, r := range resources {
			urls[i] = buildImageURL(cloudName, r)
		}

		for _, cardID := range m.cardIDs {
			if err := syncCard(ctx, pool, cardID, urls); err != nil {
				log.Printf("failed to sync card %d (folder %q): %v", cardID, m.folder, err)
				continue
			}
			results = append(results, syncResult{folder: m.folder, cardID: cardID, count: len(urls)})
		}
	}

	fmt.Println("\nSync summary:")
	fmt.Println("folder → card ID → images synced")
	for _, r := range results {
		fmt.Printf("%s → %d → %d\n", r.folder, r.cardID, r.count)
	}
}

// discoverBidBoxMappings looks up bid-box cards (IDs 18-23), slugifies their
// names, and checks whether a matching Cloudinary folder actually exists.
func discoverBidBoxMappings(ctx context.Context, pool *pgxpool.Pool, client *cloudinaryClient) []folderMapping {
	rows, err := pool.Query(ctx, "SELECT id, name FROM cards WHERE id BETWEEN $1 AND $2", bidBoxIDStart, bidBoxIDEnd)
	if err != nil {
		log.Printf("could not look up bid-box cards: %v", err)
		return nil
	}
	defer rows.Close()

	var discovered []folderMapping
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("could not scan bid-box card row: %v", err)
			continue
		}

		folder := slugify(name)
		resources, err := client.searchFolder(folder)
		if err != nil || len(resources) == 0 {
			continue
		}

		discovered = append(discovered, folderMapping{folder: folder, cardIDs: []int64{id}})
	}

	return discovered
}

func syncCard(ctx context.Context, pool *pgxpool.Pool, cardID int64, urls []string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "UPDATE cards SET image = $1, updated_at = now() WHERE id = $2", urls[0], cardID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM card_images WHERE card_id = $1", cardID); err != nil {
		return err
	}

	for i, url := range urls {
		if _, err := tx.Exec(ctx, "INSERT INTO card_images (card_id, image, sort_order) VALUES ($1, $2, $3)", cardID, url, i+1); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
