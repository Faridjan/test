package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

type ImageMeta struct {
	Src string `json:"src"`
}

func main() {
	images, err := getImagesLinksFromAdvert(context.Background(), "49635201")
	if err != nil {
		log.Fatal(err)
	}

	for _, img := range images {
		worker(img)
	}
}

func getImagesLinksFromAdvert(ctx context.Context, advertID string) ([]string, error) {
	resp, err := http.Get("https://m.krisha.kz/a/show/" + advertID)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	startCutData := strings.Split(string(body), `"photos":[{`)
	if len(startCutData) < 2 {
		log.Fatalf("Wrong advert ID")
	}

	endCutData := strings.Split(startCutData[1], `}],`)

	var listDataImages []ImageMeta

	if err := json.Unmarshal([]byte("[{"+endCutData[0]+"}]"), &listDataImages); err != nil {
		return nil, err
	}

	imagesLinks := make([]string, 0, len(listDataImages))

	for _, item := range listDataImages {
		imagesLinks = append(imagesLinks, item.Src)
	}

	return imagesLinks, nil
}

func worker(imageURL string) {
	//// Create dir...
	//if projectName == "" {
	//	projectName = time.Now().Format("2006_01_02_15_04")
	//}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Chrome Options...
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()

	opts := []chromedp.ExecAllocatorOption{
		chromedp.Headless,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("profile-directory", "Default"),
		chromedp.UserDataDir("/chromium_profile"),
		chromedp.UserAgent(getRandomUserAgent()),
	}

	ctxAllocator, cancelAllocator := chromedp.NewExecAllocator(ctx, append(chromedp.DefaultExecAllocatorOptions[:], opts...)...)
	defer cancelAllocator()

	ctxWithLog, cancelWithLog := chromedp.NewContext(ctxAllocator, chromedp.WithLogf(log.Printf))
	defer cancelWithLog()

	done := make(chan string, 1)

	chromedp.ListenTarget(ctxWithLog, func(v interface{}) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			log.Println("Listen:", ev.State)
			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.GUID
				log.Printf("state: %s, completed: %s\n", ev.State.String(), ev.GUID)
			}
		}
	})

	log.Println("Options done!")

	// Task list
	siteURL := `https://www.watermarkremover.io/ru/upload`
	imageLinkButtonSelector := `//*[@id="PasteURL__HomePage"]`
	inputImageLinkSelector := `//*[@id="modal-root"]/div/div/div[1]/div[1]/input`
	submitImageLinkSelector := `//*[@id="modal-root"]/div/div/div[1]/div[1]/button`
	downloadBtnSelector := `//*[@id="root"]/div/div[1]/div[2]/div[2]/div/div[2]/div/div/div[2]/div/div[1]/button`

	// RUN
	errRun := chromedp.Run(ctxWithLog,
		chromedp.Navigate(siteURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("- Navigate")
			return nil
		}),
		chromedp.WaitVisible(imageLinkButtonSelector),
		chromedp.Click(imageLinkButtonSelector),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("- Click URL")
			return nil
		}),
		chromedp.WaitVisible(inputImageLinkSelector),
		chromedp.SendKeys(inputImageLinkSelector, imageURL),
		chromedp.Click(submitImageLinkSelector),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("- Submit URL")
			return nil
		}),
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(wd).
			WithEventsEnabled(true),
		chromedp.WaitVisible(downloadBtnSelector),
		chromedp.Click(downloadBtnSelector),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("- Click Download")
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, err := runtime.Evaluate("window.localStorage.clear()").Do(ctx)
			if err != nil {
				return err
			}

			err = network.ClearBrowserCache().Do(ctx)
			if err != nil {
				return err
			}

			err = network.ClearBrowserCookies().Do(ctx)
			if err != nil {
				return err
			}

			return nil
		}),
	)
	if errRun != nil {
		log.Fatal(err)
	}

	log.Println("Before DONE")

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

func JsonResponse(responseWriter http.ResponseWriter, body interface{}) {
	responseWriter.WriteHeader(http.StatusOK)

	var jsonByte []byte
	var err error

	switch body.(type) {
	case proto.Message: // If body from gRPC response
		mOpt := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}

		jsonByte, err = mOpt.Marshal(body.(proto.Message))
		if err != nil {
			ErrorResponse(err, responseWriter)
			return
		}

	case map[string]interface{}: // if custom body
		jsonByte, err = json.Marshal(body)
		if err != nil {
			ErrorResponse(err, responseWriter)
			return
		}
	}

	_, err = responseWriter.Write(jsonByte)
	if err != nil {
		ErrorResponse(err, responseWriter)
		return
	}
}

func ErrorResponse(err error, w http.ResponseWriter) {
	w.WriteHeader(500)
	_, errW := w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	if errW != nil {
		log.Println("Error from ResponseWriter: " + err.Error())
		return
	}
	return
}
