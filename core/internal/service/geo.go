package service

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/probe-system/core/internal/models"
)

type GeoServiceImpl struct {
	cache      sync.Map
	httpClient *http.Client
}

func NewGeoService() *GeoServiceImpl {
	return &GeoServiceImpl{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *GeoServiceImpl) Lookup(ip string) (*models.GeoLocation, error) {
	// Check cache first
	if cached, ok := s.cache.Load(ip); ok {
		return cached.(*models.GeoLocation), nil
	}

	// Validate IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Skip private IPs
	if parsedIP.IsPrivate() || parsedIP.IsLoopback() {
		loc := &models.GeoLocation{
			Country:     "Private",
			CountryCode: "XX",
		}
		s.cache.Store(ip, loc)
		return loc, nil
	}

	// Use ip-api.com (free, no API key required)
	loc, err := s.lookupIPAPI(ip)
	if err != nil {
		// Fallback to ipinfo.io
		loc, err = s.lookupIPInfo(ip)
		if err != nil {
			return nil, err
		}
	}

	// Cache result
	s.cache.Store(ip, loc)
	return loc, nil
}

func (s *GeoServiceImpl) lookupIPAPI(ip string) (*models.GeoLocation, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"regionName"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("ip-api lookup failed")
	}

	return &models.GeoLocation{
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Region:      result.Region,
		City:        result.City,
		Latitude:    result.Lat,
		Longitude:   result.Lon,
	}, nil
}

func (s *GeoServiceImpl) lookupIPInfo(ip string) (*models.GeoLocation, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("https://ipinfo.io/%s/json", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Country string `json:"country"`
		Region  string `json:"region"`
		City    string `json:"city"`
		Loc     string `json:"loc"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var lat, lon float64
	fmt.Sscanf(result.Loc, "%f,%f", &lat, &lon)

	return &models.GeoLocation{
		Country:     result.Country,
		CountryCode: result.Country,
		Region:      result.Region,
		City:        result.City,
		Latitude:    lat,
		Longitude:   lon,
	}, nil
}

func (s *GeoServiceImpl) ClearCache() {
	s.cache = sync.Map{}
}
