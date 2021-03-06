package main

import (
  "encoding/json"
  "errors"
  "net/http"
  "net/url"
  "time"

  "github.com/appventure-nush/appventure-website/api/content"
  "io/ioutil"
  "log"
)

type API struct {
  http *http.Client
  host string
}

// Errors
var ErrorNotFound = errors.New("not found")

func (api *API) Apps() ([]content.App, error) {
  data := make(map[string][]content.App)
  params := url.Values{}
  params.Add("type", "App")
  params.Add("count", "-1")
  err := api.get("/api/contents?"+params.Encode(), &data)
  return data["data"], err
}

func (api *API) App(slug string) (content.App, error) {
  data := make(map[string][]content.App)
  params := url.Values{}
  params.Add("type", "App")
  params.Add("slug", slug)
  err := api.get("/api/content?"+params.Encode(), &data)
  if len(data["data"]) < 1 {
    return content.App{}, err
  }
  return data["data"][0], err
}

func (api *API) Projects() ([]content.Project, error) {
  data := make(map[string][]content.Project)
  params := url.Values{}
  params.Add("type", "Project")
  params.Add("count", "-1")
  err := api.get("/api/contents?"+params.Encode(), &data)
  return data["data"], err
}

func (api *API) Project(slug string) (content.Project, error) {
  data := make(map[string][]content.Project)
  params := url.Values{}
  params.Add("type", "Project")
  params.Add("slug", slug)
  err := api.get("/api/content?"+params.Encode(), &data)
  if len(data["data"]) < 1 {
    return content.Project{}, err
  }
  return data["data"][0], err
}

func (api *API) Static(filepath string) ([]byte, error) {
  log.Println("Fetching static asset:", filepath)
  data, err := api.getBytes(filepath)
  if err != nil {
    log.Println("Error fetching static asset:", err)
    return []byte{}, err
  }
  return data, err
}

func (api *API) ScreenshotByReference(ref string) (content.Screenshot, error) {
  data := make(map[string][]content.Screenshot)
  err := api.get(ref, &data)
  if len(data["data"]) < 1 {
    return content.Screenshot{}, err
  }
  return data["data"][0], err
}

// Extended types

type FeaturedApp struct {
  content.App
  content.Screenshot
}

// Extended fetching

func (api *API) FeaturedApps() ([]FeaturedApp, error) {
  all, err := api.Apps()
  if err != nil {
    return nil, err
  }
  var featured []FeaturedApp
  for _, a := range all {
    if a.Flagged("featured") {
      var s content.Screenshot
      if len(a.Screenshots) > 0 {
        s, _ = api.ScreenshotByReference(a.Screenshots[0])
      }
      featured = append(featured, FeaturedApp{
        App:        a,
        Screenshot: s,
      })
    }
  }
  return featured, err
}

func (api *API) get(path string, v interface{}) error {
  resp, err := api.http.Get(api.host + path)
  if err != nil {
    return err
  }
  if resp.StatusCode == http.StatusNotFound {
    return ErrorNotFound
  } else if resp.StatusCode == http.StatusInternalServerError {
    return ErrorNotFound
  } else if resp.StatusCode != http.StatusOK {
    return errors.New(resp.Status)
  }
  defer resp.Body.Close()
  return json.NewDecoder(resp.Body).Decode(v)
}

func (api *API) getBytes(path string) ([]byte, error) {
  resp, err := api.http.Get(api.host + path)
  if err != nil {
    return []byte{}, err
  }
  if resp.StatusCode == http.StatusNotFound {
    return []byte{}, ErrorNotFound
  } else if resp.StatusCode == http.StatusInternalServerError {
    return []byte{}, ErrorNotFound
  } else if resp.StatusCode != http.StatusOK {
    return []byte{}, errors.New(resp.Status)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  return body, err;
}

func NewAPI(host string) *API {
  return &API{
    host: host,
    http: &http.Client{
      Transport: &http.Transport{
        MaxIdleConns:       10,
        IdleConnTimeout:    30 * time.Second,
        DisableCompression: true,
      },
      Timeout: time.Second * 5,
    },
  }
}
