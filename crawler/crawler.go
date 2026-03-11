package crawler

import (
	"context"
	"errors"

	"io"
	"fmt"
	"net/url"
	"net/http"
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
	File string
}

type CrawlResult struct {
	Port  types.PortName
	Site  *url.URL
	Files []*url.URL
	Err   error
}

func NewCrawler(chanBufSize int) *Crawler {
	return &Crawler{
		in:         make(chan CrawlJob, chanBufSize),
		out:        make(chan CrawlResult, chanBufSize),
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
		if r.Site.Scheme == "http" {
			wg.Add(1)
			go func() {
				defer wg.Done()

				files, err := c.crawlHttp(r.Port, r.Site)
				c.out <- CrawlResult{
					Port:  r.Port.Name,
					Site:  r.Site,
					Files: files,
					Err:   err,
				}
			}()
			continue
		}

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

/*
func (c *Crawler) getVersionRootFromPath(port types.PortInfo, site *url.URL) (string) {
	segments := strings.Split(u.Path, "/")

	// We are looking to see if a major version is embedded in the path,
	// e.g. http://example.net/4.3/releases/file-4.3.2.zip so that when
	// we access the URL we can update this instance of the version.
	// We can handle multiple copies of the version, including different
	// levels of truncation (4.3.2, 4.3) but we'll make the assumption
	// that the first instance is the shortest (most "major" number).
	for segment := range segments {
		if len(segment) == 0 {
			continue
		}

		if strings.ContainsRune(segment, '.') && strings.HasPrefix(port.Version, segment) {
			return segment;
		}
	}

	return "";
}
*/

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

func (c *Crawler) crawlHttp(port types.PortInfo, site *url.URL) ([]*url.URL, error) {
	files := make([]*url.URL, 0)

	if c.limiter != nil {
		c.limiter.Wait(site, context.Background())
	}

	req, err := http.NewRequest("GET", site.String(), nil);

	if err != nil {
		return nil, fmt.Errorf("Error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "portscout/2")
	req.Header.Set("Content-Type", "text/html")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Error making request: %w", err)
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("Request not successful: %s", resp.Status)
	}

	defer resp.Body.Close()

	_, _ = io.ReadAll(resp.Body)

	return files, nil
}
