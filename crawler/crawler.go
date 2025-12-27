package crawler

import (
	"context"
	"errors"

	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"

	"github.com/samott/portscout2/types"
)

type Crawler struct {
	ftpTimeout time.Duration
	limiter    CrawlLimiterInterface
	in         chan CrawlJob
	out        chan CrawlResult
}

type CrawlLimiterInterface interface {
	Wait(site *url.URL, ctx context.Context)
}

type CrawlJob struct {
	Port types.PortInfo
	Site *url.URL
}

type CrawlResult struct {
	Port types.PortName
	Site *url.URL
	Files []*url.URL
	Err   error
}

func NewCrawler() *Crawler {
	return &Crawler{
		in:         make(chan CrawlJob),
		out:        make(chan CrawlResult),
		ftpTimeout: 30 * time.Second,
		limiter:    nil,
	}
}

func (c *Crawler) SetLimiter(limiter CrawlLimiterInterface) {
	c.limiter = limiter
}

func (c *Crawler) In() chan<- CrawlJob {
	return c.in
}

func (c *Crawler) Out() <-chan CrawlResult {
	return c.out
}

func (c *Crawler) Run() {
	var wg sync.WaitGroup

	for r := range c.in {
		if r.Site.Scheme == "ftp" {
			wg.Add(1)
			go func() {
				defer wg.Done()

				files, err := c.crawlFtp(r.Port, r.Site)
				c.out <- CrawlResult{
					Port:  r.Port.Name,
					Site:  r.Site,
					Files: files,
					Err:   err,
				}
			}()
			continue
		}

		// No suitable handler found
		c.out <- CrawlResult{
			Port:  r.Port.Name,
			Site:  r.Site,
			Files: nil,
			Err:   errors.New("Unhandled site scheme or format"),
		}
	}

	wg.Wait()
	close(c.out)
}

func (c *Crawler) crawlFtp(port types.PortInfo, site *url.URL) ([]*url.URL, error) {
	files := make([]*url.URL, 0)

	if c.limiter != nil {
		c.limiter.Wait(site, context.Background())
	}

	// For some reason the library doesn't use the default
	// FTP port if none is provided in the URL
	if site.Port() == "" {
		site.Host = site.Hostname() + ":21"
	}

	client, err := ftp.Dial(site.Host, ftp.DialWithTimeout(c.ftpTimeout))

	if err != nil {
		return nil, fmt.Errorf("FTP dial failed: %w", err)
	}

	err = client.Login("anonymous", "anonymous")

	if err != nil {
		return nil, fmt.Errorf("FTP login failed: %w", err)
	}

	err = client.ChangeDir(site.Path)

	if err != nil {
		return nil, fmt.Errorf("FTP cwd failed: %w", err)
	}

	entries, err := client.List(".")

	if err != nil {
		return nil, fmt.Errorf("FTP list failed: %w", err)
	}

	for _, entry := range entries {
		if entry.Type != ftp.EntryTypeFile {
			continue
		}

		fileUrl := site.JoinPath(site.String(), entry.Name)

		if err != nil {
			return nil, fmt.Errorf("URL path join failed: %w", err)
		}

		files = append(files, fileUrl)
	}

	return files, nil
}
