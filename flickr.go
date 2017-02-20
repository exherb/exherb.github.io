package main

import (
	"encoding/xml"
	"fmt"
	"github.com/spf13/hugo/hugolib"
	"github.com/spf13/hugo/parser"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const photosAPIEndpoint = "https://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=559b611ec1d4f7d1310b1b4626b98f6d&extras=date_taken,geo,url_t,url_s,url_m,url_o&photoset_id="

type Photo struct {
	Id        string  `xml:"id,attr"`
	Title     string  `xml:"title,attr"`
	Thumbnail string  `xml:"url_t,attr"`
	UrlSmall  string  `xml:"url_s,attr"`
	UrlMedium string  `xml:"url_m,attr"`
	Url       string  `xml:"url_o,attr"`
	Latitude  float32 `xml:"latitude,attr"`
	Longitude float32 `xml:"Longitude,attr"`
	Date      string  `xml:"datetaken,attr"`
}
type PhotoSet struct {
	Photos []Photo `xml:"photo"`
}
type FlickPhotosReponse struct {
	PhotoSet PhotoSet `xml:"photoset"`
}

func get_flickr_photos(flickrSetId string) []Photo {
	photos := []Photo{}

	for page := 1; ; page++ {
		apiUrl := fmt.Sprintf("%s%s&page=%d", photosAPIEndpoint, flickrSetId, page)
		response, err := http.Get(apiUrl)
		if err != nil {
			fmt.Printf("%s", err)
			return photos
		} else {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("%s", err)
				return photos
			}
			var photosReponse FlickPhotosReponse
			if err := xml.Unmarshal([]byte(body), &photosReponse); err != nil {
				fmt.Printf("%s", err)
				return photos
			}
			if len(photosReponse.PhotoSet.Photos) == 0 {
				break
			}
			photos = append(photos, photosReponse.PhotoSet.Photos...)
		}
	}
	return photos
}

func update_filick_photos(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	psr, err := parser.ReadFrom(f)
	if err != nil {
		return err
	}
	rawMetaData, err := psr.Metadata()
	if err != nil {
		return err
	}
	metaData := rawMetaData.(map[string]interface{})
	flickrsetid := metaData["flickrsetid"]
	if flickrsetid == nil {
		return nil
	}

	// add photos to front matter
	frontMatter := psr.FrontMatter()
	metaData["photos"] = get_flickr_photos(flickrsetid.(string))

	// update page
	page, _ := hugolib.NewPage(path)
	page.SetSourceMetaData(metaData, rune(frontMatter[0]))
	page.SetSourceContent(psr.Content())

	page.SaveSourceAs(path)

	return nil
}

func main() {
	if len(os.Args) > 1 && len(os.Args[1]) > 0 {
		err := update_filick_photos(os.Args[1])
		if err != nil {
			fmt.Printf("%s", err)
		}
	} else {
		filepath.Walk("./content", func(path string, f os.FileInfo, err error) error {
			fileInfo, err := os.Stat(path)
			if !fileInfo.IsDir() && filepath.Ext(path) == ".md" {
				return update_filick_photos(path)
			}
			return nil
		})
	}

}
