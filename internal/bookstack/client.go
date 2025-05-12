package bookstack

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

// Client manages crawling operations for a BookStack instance.
type Client struct {
	inst InstanceConfig
}

// NewClient creates a new Client for the given instance configuration.
func NewClient(inst InstanceConfig) *Client {
	return &Client{inst: inst}
}

// sanitizeFilename replaces invalid characters with underscores.
func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[^0-9A-Za-zÄäÖöÜüß \-_]+`)
	return re.ReplaceAllString(name, "_")
}

// htmlToMarkdown converts HTML content to Markdown using html-to-markdown converter.
func htmlToMarkdown(input string) string {
	converter := md.NewConverter("", true, nil)
	output, err := converter.ConvertString(input)
	if err != nil {
		log.Printf("Markdown conversion error: %v", err)
		return input // fallback to raw HTML
	}
	return output
}

// Crawl dispatches crawling based on the target configuration.
func (c *Client) Crawl() error {
	switch c.inst.Target.Type {
	case "book":
		return c.crawlBook(c.inst.Target.ID)
	case "shelve":
		return c.crawlShelve(c.inst.Target.ID)
	default:
		return fmt.Errorf("unsupported target type: %s", c.inst.Target.Type)
	}
}

// crawlBook fetches all pages, filters by book ID, and writes them as Markdown files.
func (c *Client) crawlBook(bookID string) error {
	log.Printf("[%s] Crawling book ID %s", c.inst.Name, bookID)
	// fetch all pages
	pagesURL := fmt.Sprintf("%s/api/pages", c.inst.BaseURL)
	body, err := c.httpGetWithAuth(pagesURL)
	if err != nil {
		return fmt.Errorf("fetching pages list: %w", err)
	}
	// parse JSON response
	type PageMeta struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		BookID    int    `json:"book_id"`
		ChapterID int    `json:"chapter_id"`
	}
	var list struct {
		Data []PageMeta `json:"data"`
	}
	if err := json.Unmarshal(body, &list); err != nil {
		return fmt.Errorf("parsing pages JSON: %w", err)
	}
	// filter pages belonging to this book
	var filtered []PageMeta
	for _, p := range list.Data {
		if fmt.Sprint(p.BookID) == bookID {
			filtered = append(filtered, p)
		}
	}
	// setup backup directory
	today := time.Now().Format("2006-01-02")
	root := filepath.Join(c.inst.BackupPath, today, fmt.Sprintf("Book_%s", bookID))
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("creating root folder: %w", err)
	}
	// group pages by chapter
	chapters := make(map[int][]PageMeta)
	for _, p := range filtered {
		chapters[p.ChapterID] = append(chapters[p.ChapterID], p)
	}
	// iterate chapters and pages
	for chapID, pages := range chapters {
		chapDir := filepath.Join(root, fmt.Sprintf("Kapitel_%d", chapID))
		if err := os.MkdirAll(chapDir, 0755); err != nil {
			log.Printf("[%s] Creating chapter folder error: %v", c.inst.Name, err)
			continue
		}
		sort.Slice(pages, func(i, j int) bool {
			return strings.ToLower(pages[i].Name) < strings.ToLower(pages[j].Name)
		})
		for idx, meta := range pages {
			detailURL := fmt.Sprintf("%s/api/pages/%d", c.inst.BaseURL, meta.ID)
			detailBody, err := c.httpGetWithAuth(detailURL)
			if err != nil {
				log.Printf("[%s] Error fetching page %d: %v", c.inst.Name, meta.ID, err)
				continue
			}
			// parse JSON page detail directly into PageDetail
			type PageDetail struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				HTML string `json:"html"`
			}
			var detail PageDetail
			if err := json.Unmarshal(detailBody, &detail); err != nil {
				log.Printf("[%s] Parsing page %d detail error: %v", c.inst.Name, meta.ID, err)
				continue
			}
			mdContent := htmlToMarkdown(detail.HTML)
			content := fmt.Sprintf("# %s\n\n%s", detail.Name, mdContent)
			filename := fmt.Sprintf("%02d_%s.md", idx+1, sanitizeFilename(detail.Name))
			path := filepath.Join(chapDir, filename)
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				log.Printf("[%s] Writing file %s error: %v", c.inst.Name, path, err)
			}
		}
	}
	log.Printf("[%s] Book crawl complete.", c.inst.Name)
	return nil
}

// crawlShelve fetches all books in the given shelve and crawls each book.
func (c *Client) crawlShelve(shelveID string) error {
	log.Printf("[%s] Crawling shelve ID %s", c.inst.Name, shelveID)
	url := fmt.Sprintf("%s/api/shelves/%s", c.inst.BaseURL, shelveID)
	body, err := c.httpGetWithAuth(url)
	if err != nil {
		return fmt.Errorf("fetching shelve: %w", err)
	}
	// parse JSON response to extract books
	type BookEntry struct {
		ID int `json:"id"`
	}
	var shelf struct {
		Books []BookEntry `json:"books"`
	}
	if err := json.Unmarshal(body, &shelf); err != nil {
		return fmt.Errorf("parsing shelve JSON: %w", err)
	}
	for _, b := range shelf.Books {
		bookID := strconv.Itoa(b.ID)
		if err := c.crawlBook(bookID); err != nil {
			log.Printf("[%s] Error crawling book %s: %v", c.inst.Name, bookID, err)
		}
	}
	log.Printf("[%s] Shelve crawl complete.", c.inst.Name)
	return nil
}
