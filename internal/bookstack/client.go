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
		for _, id := range c.inst.Target.IDs {
			if err := c.crawlBook(id); err != nil {
				log.Printf("[%s] Error crawling book ID %s: %v", c.inst.Name, id, err)
			}
		}
	case "shelve":
		for _, id := range c.inst.Target.IDs {
			if err := c.crawlShelve(id); err != nil {
				log.Printf("[%s] Error crawling shelve ID %s: %v", c.inst.Name, id, err)
			}
		}
	default:
		return fmt.Errorf("unsupported target type: %s", c.inst.Target.Type)
	}
	return nil
}

// fetchBookPages retrieves all pages for a given book ID
func (c *Client) fetchBookPages(bookID string) ([]PageMeta, error) {
	pagesURL := fmt.Sprintf("%s/api/pages", c.inst.BaseURL)
	body, err := c.httpGetWithAuth(pagesURL)
	if err != nil {
		return nil, fmt.Errorf("fetching pages list: %w", err)
	}

	var list struct {
		Data []PageMeta `json:"data"`
	}
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parsing pages JSON: %w", err)
	}

	var filtered []PageMeta
	for _, p := range list.Data {
		if fmt.Sprint(p.BookID) == bookID {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

// processPage fetches and saves a single page
func (c *Client) processPage(meta PageMeta, chapDir string, idx int) error {
	detailURL := fmt.Sprintf("%s/api/pages/%d", c.inst.BaseURL, meta.ID)
	detailBody, err := c.httpGetWithAuth(detailURL)
	if err != nil {
		return fmt.Errorf("fetching page %d: %w", meta.ID, err)
	}

	var detail PageDetail
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		return fmt.Errorf("parsing page %d detail: %w", meta.ID, err)
	}

	mdContent := htmlToMarkdown(detail.HTML)
	content := fmt.Sprintf("# %s\n\n%s", detail.Name, mdContent)
	filename := fmt.Sprintf("%02d_%s.md", idx+1, sanitizeFilename(detail.Name))
	path := filepath.Join(chapDir, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

// crawlBook fetches all pages, filters by book ID, and writes them as Markdown files.
func (c *Client) crawlBook(bookID string, shelveID ...string) error {
	log.Printf("[%s] Crawling book ID %s", c.inst.Name, bookID)

	pages, err := c.fetchBookPages(bookID)
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")
	var root string
	if len(shelveID) > 0 {
		root = filepath.Join(c.inst.BackupPath, today, fmt.Sprintf("shelves_%s", shelveID[0]), fmt.Sprintf("book_%s", bookID))
	} else {
		root = filepath.Join(c.inst.BackupPath, today, fmt.Sprintf("book_%s", bookID))
	}
	log.Printf("[%s] Creating root directory: %s", c.inst.Name, root)
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("creating root folder: %w", err)
	}

	chapters := make(map[int][]PageMeta)
	for _, p := range pages {
		chapters[p.ChapterID] = append(chapters[p.ChapterID], p)
	}

	for chapID, pages := range chapters {
		chapDir := filepath.Join(root, fmt.Sprintf("Kapitel_%d", chapID))
		log.Printf("[%s] Creating chapter directory: %s", c.inst.Name, chapDir)
		if err := os.MkdirAll(chapDir, 0755); err != nil {
			log.Printf("[%s] Creating chapter folder error: %v", c.inst.Name, err)
			continue
		}

		sort.Slice(pages, func(i, j int) bool {
			return strings.ToLower(pages[i].Name) < strings.ToLower(pages[j].Name)
		})

		for idx, meta := range pages {
			if err := c.processPage(meta, chapDir, idx); err != nil {
				log.Printf("[%s] Error processing page %d: %v", c.inst.Name, meta.ID, err)
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
		if err := c.crawlBook(bookID, shelveID); err != nil {
			log.Printf("[%s] Error crawling book %s: %v", c.inst.Name, bookID, err)
		}
	}
	log.Printf("[%s] Shelve crawl complete.", c.inst.Name)
	return nil
}
