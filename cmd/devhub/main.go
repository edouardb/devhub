package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	"github.com/renstrom/fuzzysearch/fuzzy"
	"github.com/shazow/memoizer"

	"github.com/scaleway/devhub/pkg/image"
	"github.com/scaleway/devhub/pkg/manifest"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

var cache Cache
var memoize memoizer.Memoize
var httpMutex sync.Mutex
var badgeMutex sync.Mutex

func init() {
	memoize = memoizer.Memoize{memoizer.NewMemoryCache()}
}

type Cache struct {
	Mapping struct {
		Images []*ImageMapping `json:"images"`
	} `json:"mapping"`
	Manifest *scwManifest.Manifest `json:"manifest"`
	Api      struct {
		Images      *[]api.ScalewayImage      `json:"api_images"`
		Bootscripts *[]api.ScalewayBootscript `json:"api_bootscripts"`
	} `json:"api"`
}

func NewCache() Cache {
	return Cache{}
}

func (c *Cache) GetImageByName(name string) []*ImageMapping {
	found := []*ImageMapping{}
	for _, image := range c.Mapping.Images {
		if image.MatchName(name) {
			found = append(found, image)
		}
	}
	return found
}

func (c *Cache) MapImages() {
	// FIXME: add mutex
	if c.Manifest == nil || c.Api.Images == nil {
		return
	}

	c.Mapping.Images = make([]*ImageMapping, 0)

	logrus.Infof("Mapping images")
	for _, manifestImage := range c.Manifest.Images {
		imageMapping := ImageMapping{
			ManifestName: manifestImage.FullName(),
		}
		imageMapping.Objects.Manifest = manifestImage
		manifestImageName := ImageCodeName(manifestImage.Name)
		for _, apiImage := range *c.Api.Images {
			apiImageName := ImageCodeName(apiImage.Name)
			if rankMatch := fuzzy.RankMatch(manifestImageName, apiImageName); rankMatch > -1 {
				imageMapping.ApiUUID = apiImage.Identifier
				imageMapping.RankMatch = rankMatch
				imageMapping.Found++
				imageMapping.Objects.Api = &apiImage
			}
		}
		c.Mapping.Images = append(c.Mapping.Images, &imageMapping)
	}
	logrus.Infof("Images mapped")
}

type ImageMapping struct {
	ApiUUID      string `json:"api_uuid"`
	ManifestName string `json:"manifest_name"`
	RankMatch    int    `json:"rank_match"`
	Found        int    `json:"found"`

	Objects struct {
		Api      *api.ScalewayImage `json:"api"`
		Manifest *scwImage.Image    `json:"manifest"`
	} `json:"objects"`
}

func (i *ImageMapping) MatchName(input string) bool {
	if input == i.ApiUUID {
		return true
	}

	input = ImageCodeName(input)

	if fuzzy.RankMatch(input, ImageCodeName(i.ManifestName)) > -1 {
		return true
	}

	for _, tag := range i.Objects.Manifest.Tags {
		nameWithTag := ImageCodeName(fmt.Sprintf("%s-%s", i.Objects.Manifest.Name, tag))
		if fuzzy.RankMatch(input, nameWithTag) > -1 {
			return true
		}
	}

	return false
}

func ImageCodeName(inputName string) string {
	name := strings.ToLower(inputName)
	name = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`--+`).ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return name
}

func main() {
	router := gin.Default()

	router.StaticFile("/", "./static/index.html")
	router.Static("/static", "./static")
	router.Static("/bower_components", "./bower_components")
	// FIXME: favicon

	router.GET("/api/images", imagesEndpoint)
	router.GET("/api/images/:name", imageEndpoint)
	router.GET("/api/images/:name/dockerfile", imageDockerfileEndpoint)
	router.GET("/api/images/:name/makefile", imageMakefileEndpoint)

	router.GET("/api/bootscripts", bootscriptsEndpoint)

	router.GET("/api/cache", cacheEndpoint)

	router.GET("/images/:name/new-server", newServerEndpoint)

	router.GET("/badges/images/:name/scw-build.svg", badgeImageScwBuild)
	// router.GET("/images/:name/badge", imageBadgeEndpoint)

	Api, err := api.NewScalewayAPI("https://api.scaleway.com", "", os.Getenv("SCALEWAY_ORGANIZATION"), os.Getenv("SCALEWAY_TOKEN"))
	if err != nil {
		logrus.Fatalf("Failed to initialize Scaleway Api: %v", err)
	}

	go updateManifestCron(&cache)
	go updateScwApiImages(Api, &cache)
	go updateScwApiBootscripts(Api, &cache)

	router.Run(":4242")
}

func httpGetContent(url string) (string, error) {
	httpMutex.Lock()
	defer httpMutex.Unlock()
	logrus.Warnf("Fetching HTTP content: %s", url)
	request := gorequest.New()
	_, body, errs := request.Get(url).End()
	if len(errs) > 0 {
		// FIXME: return all the errs
		return "", errs[0]
	}
	return body, nil
}

func getBadge(left, right, color string) (string, error) {
	badgeMutex.Lock()
	defer badgeMutex.Unlock()
	left = strings.Replace(left, "-", "--", -1)
	right = strings.Replace(right, "-", "--", -1)
	left = strings.Replace(left, " ", "_", -1)
	right = strings.Replace(right, " ", "_", -1)
	url := fmt.Sprintf("https://img.shields.io/badge/%s-%s-%s.svg", left, right, color)
	body, err := memoize.Call(httpGetContent, url)
	if err != nil {
		return errBadge(left, fmt.Errorf("http error")), err
	}
	return body.(string), nil
}

func errBadge(left string, err error) string {
	logrus.Warnf("Failed to get badge: %v", err)

	// FIXME: remove the switch and compute the left width automatically
	switch left {
	case "build":
		return `<svg xmlns="http://www.w3.org/2000/svg" width="115" height="20"><linearGradient id="b" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="a"><rect width="115" height="20" rx="3" fill="#fff"/></mask><g mask="url(#a)"><path fill="#555" d="M0 0h37v20H0z"/><path fill="#9f9f9f" d="M37 0h78v20H37z"/><path fill="url(#b)" d="M0 0h115v20H0z"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="18.5" y="15" fill="#010101" fill-opacity=".3">build</text><text x="18.5" y="14">build</text><text x="75" y="15" fill="#010101" fill-opacity=".3">inaccessible</text><text x="75" y="14">inaccessible</text></g></svg>`
	default:
		return strings.Replace(`<svg xmlns="http://www.w3.org/2000/svg" width="133" height="20"><linearGradient id="b" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="a"><rect width="133" height="20" rx="3" fill="#fff"/></mask><g mask="url(#a)"><path fill="#555" d="M0 0h55v20H0z"/><path fill="#9f9f9f" d="M55 0h78v20H55z"/><path fill="url(#b)" d="M0 0h133v20H0z"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="27.5" y="15" fill="#010101" fill-opacity=".3">{{ .Left }}</text><text x="27.5" y="14">{{ .Left }}</text><text x="93" y="15" fill="#010101" fill-opacity=".3">inaccessible</text><text x="93" y="14">inaccessible</text></g></svg>`, "{{ .Left }}", left, -1)
	}
}

func badgeImageScwBuild(c *gin.Context) {
	name := c.Param("name")
	images := cache.GetImageByName(name)
	left := "build"
	c.Header("Content-Type", "image/svg+xml;charset=utf-8")
	switch len(images) {
	case 0:
		c.String(http.StatusNotFound, errBadge(left, fmt.Errorf("no such image")))
	case 1:
		image := images[0]
		if image.Objects.Api == nil {
			c.String(http.StatusInternalServerError, errBadge(left, fmt.Errorf("invalid resource")))
			return
		}
		creationDate, err := time.Parse(time.RFC3339, image.Objects.Api.CreationDate)
		if err != nil {
			c.String(http.StatusInternalServerError, errBadge(left, fmt.Errorf("invalid-date")))
			return
		}

		humanTime := humanize.Time(creationDate)
		humanTime = strings.Replace(humanTime, " ago", "", -1)

		badge, err := getBadge(left, humanTime, "green")
		if err != nil {
			c.String(http.StatusInternalServerError, badge)
			return
		}
		c.String(http.StatusOK, badge)
	default:
		c.String(http.StatusNotFound, errBadge(left, fmt.Errorf("ambiguous name")))
	}
}

func cacheEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"cache": cache,
	})
}

func imagesEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"images": cache.Mapping.Images,
	})
}

func bootscriptsEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"bootscripts": cache.Api.Bootscripts,
	})
}

func imageDockerfileEndpoint(c *gin.Context) {
	name := c.Param("name")
	image := cache.Manifest.Images[name]
	dockerfile, err := image.GetDockerfile()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("%v", err),
		})
	}
	c.String(http.StatusOK, dockerfile)
}

func imageMakefileEndpoint(c *gin.Context) {
	name := c.Param("name")
	image := cache.Manifest.Images[name]
	makefile, err := image.GetMakefile()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("%v", err),
		})
	}
	c.String(http.StatusOK, makefile)
}

func imageEndpoint(c *gin.Context) {
	name := c.Param("name")
	images := cache.GetImageByName(name)
	switch len(images) {
	case 0:
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No such image",
		})
	case 1:
		c.JSON(http.StatusOK, gin.H{
			"image": images[0],
		})
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Too much images are matching your request",
			"images": images,
		})
	}
}

func newServerEndpoint(c *gin.Context) {
	name := c.Param("name")
	images := cache.GetImageByName(name)
	switch len(images) {
	case 0:
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No such image",
		})
	case 1:
		c.Redirect(http.StatusFound, fmt.Sprintf("https://cloud.scaleway.com/#/servers/new?image=%s", images[0].ApiUUID))
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Too much images are matching your request",
			"images": images,
		})
	}
}

func updateScwApiImages(Api *api.ScalewayAPI, cache *Cache) {
	for {
		logrus.Infof("Fetching images from the Api...")
		images, err := Api.GetImages()
		if err != nil {
			logrus.Errorf("Failed to retrieve images list from the Api: %v", err)
		} else {
			cache.Api.Images = images
			logrus.Infof("Images fetched: %d images", len(*images))
			cache.MapImages()
		}
		time.Sleep(5 * time.Minute)
	}
}

func updateScwApiBootscripts(Api *api.ScalewayAPI, cache *Cache) {
	for {
		logrus.Infof("Fetching bootscripts from the Api...")
		bootscripts, err := Api.GetBootscripts()
		if err != nil {
			logrus.Errorf("Failed to retrieve bootscripts list from the Api: %v", err)
		} else {
			cache.Api.Bootscripts = bootscripts
			logrus.Infof("Bootscripts fetched: %d bootscripts", len(*bootscripts))
		}
		time.Sleep(5 * time.Minute)
	}
}

func updateManifestCron(cache *Cache) {
	for {
		logrus.Infof("Fetching manifest...")
		manifest, err := scwManifest.GetManifest()
		if err != nil {
			logrus.Errorf("Cannot get manifest: %v", err)
		} else {
			cache.Manifest = manifest
			logrus.Infof("Manifest fetched: %d images", len(manifest.Images))
			cache.MapImages()
		}
		time.Sleep(5 * time.Minute)
	}
}
