package main

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/gocolly/colly"
	"github.com/igorkichuk/scrapper/internal/pkg/fatalfuncs"
	"github.com/igorkichuk/scrapper/internal/pkg/match"
	"github.com/jasonwinn/geocoder"
	"github.com/joho/godotenv"
	"net/http"
)

const (
	biCtxKey = "branchInfo"
)

var (
	errBranchLink        = errors.New("link for the branch wasn't found")
	errBranchReqCreation = errors.New("failed request creation for the branch")
	errGettingContext    = errors.New("failed getting context")
)

type branchInfo struct {
	Name        string   `json:"name"`
	Addr        string   `json:"addr"`
	Phone       string   `json:"phone"`
	Location    location `json:"location"`
	StaffEmails []string `json:"staff_emails"`
}

type location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type config struct {
	ResultFileName string `env:"RESULT_FILE_NAME,required"`
	BeginningUrl   string `env:"BEGINNING_URL,required"`
}

func main() {
	// reading configs
	cfg := config{}
	_ = godotenv.Load()
	err := env.Parse(&cfg)
	fatalfuncs.CheckErr(err)

	// variables initialization
	baseC := colly.NewCollector()
	staffC := colly.NewCollector()
	var branchesInfo []branchInfo

	// scrapping branches info at beginning page and initialize scrapping branch staff emails
	baseC.OnHTML(".node", func(e *colly.HTMLElement) {
		addr := e.ChildText(".field-location-direction")
		bi := branchInfo{
			Name:     e.ChildText(".location-card-header > h2 > span"),
			Addr:     addr,
			Phone:    e.ChildText(".wrapper-field-location-phone a"),
			Location: getLocation(addr),
		}

		branchUrl, err := getBranchUrl(e)
		if err != nil {
			fmt.Println(err, bi.Name)
			branchesInfo = append(branchesInfo, bi)
			return
		}

		ctx := colly.NewContext()
		ctx.Put(biCtxKey, bi)
		err = staffC.Request(http.MethodGet, branchUrl+"/about", nil, ctx, nil)
		if err != nil {
			fmt.Println(errBranchReqCreation, bi)
			branchesInfo = append(branchesInfo, bi)
		}
	})

	staffC.OnHTML("body", parseStaffCallback)

	// receiving crapped info and adding it ti branchesInfo slice
	staffC.OnScraped(func(r *colly.Response) {
		if r.StatusCode != http.StatusOK {
			return
		}

		bi, err := getBranchInfoFromCtx(r.Ctx)
		if err != nil {
			fmt.Println(err)
			return
		}

		branchesInfo = append(branchesInfo, bi)
	})

	_ = baseC.Visit(cfg.BeginningUrl)
	fatalfuncs.SaveJsonToFile(cfg.ResultFileName, branchesInfo)
	fmt.Println("Script has finished.")
}

// parseStaffCallback is callback function for parsing emails of staff and putting them into response context.
func parseStaffCallback(e *colly.HTMLElement) {
	var emails []string
	e.ForEach(".block-description--wrapper .left-col p a", func(i int, el *colly.HTMLElement) {
		ok, _ := match.Email(el.Text)
		if ok {
			emails = append(emails, el.Text)
		}
	})

	bi, err := getBranchInfoFromCtx(e.Request.Ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	bi.StaffEmails = emails
	e.Response.Ctx.Put(biCtxKey, bi)
}

// getBranchUrl parses url from branch html element.
func getBranchUrl(e *colly.HTMLElement) (string, error) {
	branchUrl, ok := e.DOM.Find("span > a").Attr("href")
	if !ok {
		return "", errBranchLink
	}

	return e.Request.AbsoluteURL(branchUrl), nil
}

// getLocation uses Google Geocode API and returns location.
// In case when Google Geocode API returns an error, this function returns an empty location structure.
func getLocation(addr string) location {
	location := location{}
	lat, lng, err := geocoder.Geocode(addr)
	if err == nil {
		location.Latitude = lat
		location.Longitude = lng
	}
	return location
}

// getBranchInfoFromCtx returns branchInfo from context received as parameter.
func getBranchInfoFromCtx(ctx *colly.Context) (branchInfo, error) {
	bi, ok := ctx.GetAny(biCtxKey).(branchInfo)
	if !ok {
		return branchInfo{}, fmt.Errorf("%w: %s", errGettingContext, biCtxKey)
	}
	return bi, nil
}
