package mlclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) POST(path string, body interface{}, result interface{}) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("ml service call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ml service error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode ml response: %w", err)
	}

	return nil
}

type KatalogEkstrak struct {
	Hari          []string `json:"hari"`
	Jam           []string `json:"jam"`
	MataPelajaran []string `json:"mata_pelajaran"`
	Kelas         []string `json:"kelas"`
	Guru          []string `json:"guru"`
	Ruangan       []string `json:"ruangan"`
}

type BarisEkstrak struct {
	Hari          string `json:"hari"`
	Jam           string `json:"jam"`
	MataPelajaran string `json:"mata_pelajaran"`
	Kelas         string `json:"kelas"`
	Guru          string `json:"guru"`
	Ruangan       string `json:"ruangan"`
}

// EkstrakJadwal mengirim berkas ke layanan ML. Timeout terpisah dari POST JSON 30 detik.
func (c *Client) EkstrakJadwal(namaBerkas string, isi []byte, katalog KatalogEkstrak) ([]BarisEkstrak, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	bagian, err := w.CreateFormFile("berkas", namaBerkas)
	if err != nil {
		return nil, fmt.Errorf("siapkan berkas ekstraksi: %w", err)
	}
	if _, err := bagian.Write(isi); err != nil {
		return nil, fmt.Errorf("tulis berkas ekstraksi: %w", err)
	}
	katalogJSON, err := json.Marshal(katalog)
	if err != nil {
		return nil, fmt.Errorf("katalog ekstraksi: %w", err)
	}
	if err := w.WriteField("katalog", string(katalogJSON)); err != nil {
		return nil, fmt.Errorf("katalog ekstraksi: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("tutup berkas ekstraksi: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/jadwal/ekstrak", &buf)
	if err != nil {
		return nil, fmt.Errorf("permintaan ekstraksi: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	klien := &http.Client{Timeout: 120 * time.Second}
	resp, err := klien.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ml service call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 800))
		return nil, fmt.Errorf("ml service error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var out struct {
		Baris []BarisEkstrak `json:"baris"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode ml response: %w", err)
	}
	if out.Baris == nil {
		out.Baris = []BarisEkstrak{}
	}
	return out.Baris, nil
}
