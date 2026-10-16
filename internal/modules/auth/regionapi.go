package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// The browser must not call the region master-data provider directly: it can
// be blocked by corporate web filters and exposes the provider's availability
// to every visitor. The API fetches it server-side, caches it, and serves the
// same data (and validates submitted codes) from that cache.
const (
	defaultRegionBaseURL = "https://www.emsifa.com/api-wilayah-indonesia/v2"
	regionCacheTTL       = 24 * time.Hour
	maxRegionBodyBytes   = 2 << 20
)

var provinceCodePattern = regexp.MustCompile(`^[0-9]{2}$`)

type regionItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type regionCacheEntry struct {
	items     []regionItem
	fetchedAt time.Time
}

type regionCache struct {
	mu      sync.Mutex
	entries map[string]regionCacheEntry
}

func (s *Service) regionBase() string {
	if s.regionBaseURL != "" {
		return s.regionBaseURL
	}
	return defaultRegionBaseURL
}

// regionItems returns provinces ("provinces.json") or the regencies of one
// province ("regencies/34.json"). A stale cached copy is served if the
// provider is unreachable.
func (s *Service) regionItems(ctx context.Context, path string, idPattern *regexp.Regexp) ([]regionItem, error) {
	s.regions.mu.Lock()
	entry, cached := s.regions.entries[path]
	s.regions.mu.Unlock()
	if cached && s.now().Sub(entry.fetchedAt) < regionCacheTTL {
		return entry.items, nil
	}

	items, err := s.fetchRegionItems(ctx, path, idPattern)
	if err != nil {
		if cached {
			return entry.items, nil
		}
		return nil, err
	}
	s.regions.mu.Lock()
	if s.regions.entries == nil {
		s.regions.entries = make(map[string]regionCacheEntry)
	}
	s.regions.entries[path] = regionCacheEntry{items: items, fetchedAt: s.now()}
	s.regions.mu.Unlock()
	return items, nil
}

func (s *Service) fetchRegionItems(ctx context.Context, path string, idPattern *regexp.Regexp) ([]regionItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.regionBase()+"/"+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create region request: %w", err)
	}
	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch region data: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("region provider returned HTTP %d", response.StatusCode)
	}
	var body struct {
		Data []regionItem `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxRegionBodyBytes)).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode region data: %w", err)
	}
	items := make([]regionItem, 0, len(body.Data))
	for _, item := range body.Data {
		if idPattern.MatchString(item.ID) && item.Name != "" && len(item.Name) <= 120 {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("region provider returned no usable items")
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

func (s *Service) provinces(ctx context.Context) ([]regionItem, error) {
	return s.regionItems(ctx, "provinces.json", provinceCodePattern)
}

func (s *Service) regencies(ctx context.Context, provinceCode string) ([]regionItem, error) {
	return s.regionItems(ctx, "regencies/"+provinceCode+".json", domesticCityCodePattern)
}

func (h *Handler) registerRegionRoutes(router chi.Router) {
	router.Get("/api/v1/regions/provinces", h.listProvinces)
	router.Get("/api/v1/regions/provinces/{code}/regencies", h.listRegencies)
}

func (h *Handler) listProvinces(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.provinces(r.Context())
	h.writeRegions(w, items, err, "list_provinces")
}

func (h *Handler) listRegencies(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if !provinceCodePattern.MatchString(code) {
		writeError(w, http.StatusNotFound, "province not found")
		return
	}
	items, err := h.service.regencies(r.Context(), code)
	h.writeRegions(w, items, err, "list_regencies")
}

func (h *Handler) writeRegions(w http.ResponseWriter, items []regionItem, err error, operation string) {
	if err != nil {
		h.reportFailure(operation, err)
		writeError(w, http.StatusServiceUnavailable, "region data is temporarily unavailable")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "content": map[string]any{"items": items}})
}
