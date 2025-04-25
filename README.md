# 📚 BookStack Crawler

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/bookstack-crawler.svg)](https://pkg.go.dev/github.com/yourusername/bookstack-crawler)

> A Go package to **crawl** and **backup** content from [BookStack](https://www.bookstackapp.com/) instances into **Markdown files** — organized, scheduled, and easily accessible.

---

## ✨ Features

- 🖋 Backup BookStack pages as Markdown files
- 📂 Organize backups by books, chapters, and pages
- ⏰ Cron-based automated crawling
- 🔐 Secure API access using Token ID and Secret
- 📚 Support for both **Book** and **Shelve** targets

---

## 📦 Installation

```bash
go get github.com/Flexible-Universe/bookstack-crawler
```

Install required dependencies:

```bash
go get gopkg.in/yaml.v2
go get github.com/JohannesKaufmann/html-to-markdown
go get github.com/robfig/cron/v3
```

---

## ⚙️ Configuration

Create a `config.yaml` file in your project directory.

Example:

```yaml
instances:
  - name: "MyBookStack"
    base_url: "https://bookstack.example.com"
    token_id: "your-token-id"
    token_secret: "your-token-secret"
    backup_path: "./backups"
    schedule: "@daily"
    target:
      type: "book"    # or "shelve"
      id: "1"
```

| Field         | Description                                               |
| ------------- | ---------------------------------------------------------- |
| `name`        | Human-readable name for the instance.                      |
| `base_url`    | Base URL of your BookStack instance.                        |
| `token_id`    | API Token ID for authentication.                            |
| `token_secret`| API Token Secret for authentication.                        |
| `backup_path` | Local path where backups will be stored.                    |
| `schedule`    | Cron expression for automated backups (e.g., `@daily`).     |
| `target.type` | `"book"` or `"shelve"`.                                      |
| `target.id`   | ID of the book or shelve you want to backup.                 |

---

## 🚀 Usage

### Start the scheduler and run automatically:

```go
package main

import (
	"log"
	"github.com/Flexible-Universe/bookstack-crawler/bookstack"
)

func main() {
	config, err := bookstack.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	scheduler, err := bookstack.Scheduler(config)
	if err != nil {
		log.Fatalf("Error starting scheduler: %v", err)
	}

	// Keep the application alive
	select {}
}
```

### Manually trigger a backup:

```go
client := bookstack.NewClient(config.Instances[0])
if err := client.Crawl(); err != nil {
	log.Fatalf("Error during manual crawl: %v", err)
}
```

---

## 📝 License

This project is licensed under the [MIT License](https://github.com/Flexible-Universe/bookstack-crawler?tab=MIT-1-ov-file).

---
