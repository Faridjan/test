package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/target"

	"github.com/chromedp/cdproto/page"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

var mu sync.Mutex

func main() {

	http.HandleFunc("/remove", RemoveByLink)

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Println(err.Error())
	}
}

func RemoveByLink(w http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		log.Println(err)
		return
	}

	m := map[string]string{}
	for k, v := range request.Form {
		m[k] = v[0]
	}

	log.Println("Params:", m)

	// Create dir...
	projectName := m["project_name"]
	if projectName == "" {
		projectName = time.Now().Format("2006_01_02_15_04")
	}

	ex, err := os.Executable()
	if err != nil {
		log.Println(err)
		return
	}
	exPath := filepath.Dir(ex)

	path := exPath + "/temp/" + projectName

	ctx := context.Background()

	opts := []chromedp.ExecAllocatorOption{
		chromedp.Headless,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("profile-directory", "Default"),
		chromedp.UserDataDir(exPath + "/chromium_profile"),
		chromedp.UserAgent(getRandomUserAgent()),
	}

	ctxAllocator, cancelAllocator := chromedp.NewExecAllocator(ctx, append(chromedp.DefaultExecAllocatorOptions[:], opts...)...)
	defer cancelAllocator()

	ctxWithLog, cancelWithLog := chromedp.NewContext(ctxAllocator, chromedp.WithLogf(log.Printf))
	defer cancelWithLog()

	done := make(chan string, len(m["image_link"]))

	chromedp.ListenTarget(ctxWithLog, func(v interface{}) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			log.Println("Listen:", ev.State)
			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.GUID
				log.Printf("state: %s, completed: %s\n", ev.State.String(), ev.GUID)
			}
		}
	})

	var errChan = make(chan error, 1)

	chromedp.ListenTarget(ctxWithLog, func(ev interface{}) {
		switch ev := ev.(type) {
		case *cdpruntime.EventConsoleAPICalled:
			if ev.Type == "error" {
				log.Println("target event: received error", ev)
			}

		case *cdpruntime.EventExceptionThrown:
			log.Println("target event: received exception", ev)

		case *target.EventTargetCrashed:
			log.Println("target event: received crashed event", ev)

		case *browser.EventDownloadProgress:
			if ev.State == browser.DownloadProgressStateCanceled {
				errChan <- errors.New(browser.DownloadProgressStateCanceled.String())
			}
		}
	})

	mu.Lock()

	log.Println("Options done!")

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Println("Create dir:", path)
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		log.Println(err)
		return
	}

	mu.Unlock()

	// Task list
	siteURL := `https://www.watermarkremover.io/ru/upload`
	imageLinkButtonSelector := `//*[@id="PasteURL__HomePage"]`
	inputImageLinkSelector := `//*[@id="modal-root"]/div/div/div[1]/div[1]/input`

	submitImageLinkSelector := `//*[@id="modal-root"]/div/div/div[1]/div[1]/button`
	downloadBtnSelector := `//*[@id="root"]/div/div[1]/div[2]/div[2]/div/div[2]/div/div/div[2]/div/div[1]/button`

	var buf []byte

	// RUN
	err = chromedp.Run(ctxWithLog,
		chromedp.Navigate(siteURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("- Navigate")
			chromedp.FullScreenshot(&buf, 90)
			if err := os.WriteFile("fullScreenshot.png", buf, 0o644); err != nil {
				log.Fatal(err)
			}
			return nil
		}),
		chromedp.WaitVisible(imageLinkButtonSelector),
		chromedp.Click(imageLinkButtonSelector),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("- Click URL")
			chromedp.FullScreenshot(&buf, 90)
			if err := os.WriteFile("fullScreenshot2.png", buf, 0o644); err != nil {
				log.Fatal(err)
			}
			return nil
		}),
		chromedp.WaitVisible(inputImageLinkSelector),
		chromedp.SendKeys(inputImageLinkSelector, m["image_link"]),
		chromedp.Click(submitImageLinkSelector),
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromedp.FullScreenshot(&buf, 90)
			if err := os.WriteFile("fullScreenshot3.png", buf, 0o644); err != nil {
				log.Fatal(err)
			}
			log.Println("- Send image")
			return nil
		}),
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(path).
			WithEventsEnabled(true),
		page.SetDownloadBehavior(page.SetDownloadBehaviorBehaviorAllow).WithDownloadPath(path),
		chromedp.WaitVisible(downloadBtnSelector),
		chromedp.Click(downloadBtnSelector),
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromedp.FullScreenshot(&buf, 90)
			if err := os.WriteFile("fullScreenshot4.png", buf, 0o644); err != nil {
				log.Fatal(err)
			}
			log.Println("- Click download")
			return nil
		}),
	)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Before DONE")

	time.Sleep(3 * time.Second)
	guid := <-done

	log.Println("DONE!", guid)
}

func getRandomUserAgent() string {
	listUserAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:105.0) Gecko/20100101 Firefox/105.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:106.0) Gecko/20100101 Firefox/106.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:105.0) Gecko/20100101 Firefox/105.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; rv:105.0) Gecko/20100101 Firefox/105.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:105.0) Gecko/20100101 Firefox/105.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:106.0) Gecko/20100101 Firefox/106.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:105.0) Gecko/20100101 Firefox/105.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 Edg/106.0.1370.52",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 Edg/106.0.1370.42",
		"Mozilla/5.0 (Windows NT 10.0; rv:106.0) Gecko/20100101 Firefox/106.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:102.0) Gecko/20100101 Firefox/102.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:106.0) Gecko/20100101 Firefox/106.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 Edg/106.0.1370.47",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 Edg/106.0.1370.37",
		"Mozilla/5.0 (X11; Linux x86_64; rv:103.0) Gecko/20100101 Firefox/103.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 Edg/105.0.1343.53",
		"Mozilla/5.0 (X11; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 OPR/91.0.4516.77",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 Edg/106.0.1370.34",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:106.0) Gecko/20100101 Firefox/106.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:104.0) Gecko/20100101 Firefox/104.0",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:104.0) Gecko/20100101 Firefox/104.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; rv:91.0) Gecko/20100101 Firefox/91.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edg/107.0.1418.24",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 Edg/105.0.1343.42",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:105.0) Gecko/20100101 Firefox/105.0",
		"Mozilla/5.0 (Windows NT 10.0; rv:102.0) Gecko/20100101 Firefox/102.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 OPR/91.0.4516.65",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:91.0) Gecko/20100101 Firefox/91.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:104.0) Gecko/20100101 Firefox/104.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.102 Safari/537.36 OPR/90.0.4480.117",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.102 Safari/537.36 OPR/90.0.4480.84",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:100.0) Gecko/20100101 Firefox/100.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko)",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
	}
	randUserAgent := rand.Intn((len(listUserAgents)-1)-0) + 0

	return listUserAgents[randUserAgent]
}
