package main

import (
	"fmt"

	"encoding/json"
	"github.com/mrjones/oauth"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
)

var (
	consumerKey       = "hnVOVwYCMc5LFwLZw1UIzwZlN"
	consumerSecret    = "X3xwZPVaQI37Kh6LEJKCNklM44Fp3uYLJ0WpuZpu6OKBwhUuVO"
	accessToken       = "173763300-w5PtHuM0j8EGIwcJALHSxkzgkt5Z8IMmWZv0feVk"
	accessTokenSecret = "HJ10by4uP5oE3GC4oebWIb8BKFi4AA7amDTqEsEyikbCP"
)

func main() {

	http.HandleFunc("/", requestHandler)
	http.ListenAndServe(":9000", nil)

	//fasthttp.ListenAndServe(":9000", requestHandler)
	/*if err := fasthttp.ListenAndServe(*addr, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}*/
}

type GoogleResponse struct {
	Items []Item `json:"items"`
}

type Item struct {
	Snippet string `json:"snippet"`
}

type DuckDuckResponse struct {
	RelatedTopics []RelatedTopic `json:"RelatedTopics"`
}

type RelatedTopic struct {
	Text string `json:"Text"`
}

type twitterResponse struct {
	Statuses []statuses `json:"statuses"`
}

type statuses struct {
	Text string `json:"text"`
}

type FinalResult struct {
	Result Result `json:"result"`
}

type Result struct {
	Google     Google     `json:"google"`
	Twitter    Twitter    `json:"twitter"`
	DuckDuckGo DuckDuckGo `json:"duckduckgo"`
}

type Google struct {
	Url  string `json:"url"`
	Text string `json:"text"`
}

type Twitter struct {
	Url  string `json:"url"`
	Text string `json:"text"`
}

type DuckDuckGo struct {
	Url  string `json:"url"`
	Text string `json:"text"`
}

func requestHandler(res http.ResponseWriter, req *http.Request) {

	googleAPI := "https://www.googleapis.com/customsearch/v1?key=AIzaSyClAIcUsDXqqHxtbuxa1LNivK7twgYKvyk&cx=015608138469747526363:dvommr-vmpa"
	duckDuckGoAPI := "http://api.duckuckgo.com/?format=json"
	twitterAPI := "https://api.twitter.com/1.1/search/tweets.json"

	q := req.URL.Query().Get("q")

	chGoogle := make(chan string)
	chTwitter := make(chan string)
	chDuckDuckGo := make(chan string)
	go googleSearch(chGoogle, q, &googleAPI)
	go twitterTweet(chTwitter, q, &twitterAPI)
	go duckDuck(chDuckDuckGo, q, &duckDuckGoAPI)
	var finalResult FinalResult
	twitterDone := false
	googleDone := false
	duckDuckGoDone := false
	for {
		select {
		case r := <-chDuckDuckGo:
			finalResult.Result.DuckDuckGo.Text = r
			finalResult.Result.DuckDuckGo.Url = duckDuckGoAPI
			duckDuckGoDone = true
		case r := <-chGoogle:
			finalResult.Result.Google.Text = r
			finalResult.Result.Google.Url = googleAPI
			googleDone = true
		case r := <-chTwitter:
			finalResult.Result.Twitter.Text = r
			finalResult.Result.Twitter.Url = twitterAPI
			twitterDone = true

		}
		if duckDuckGoDone && twitterDone && googleDone {
			break
		}
	}


	bodyResult, err := json.Marshal(&finalResult)
	if err != nil {
		res.Write([]byte(err.Error()) )
	}
	res.Write(bodyResult)

}

func duckDuck(chDuckDuckGo chan string, q string, duckDuckGoAPI *string) {

	*duckDuckGoAPI = fmt.Sprint(*duckDuckGoAPI + "&q=" + q)

	_, bodyDuck, err := fasthttp.Get(nil, *duckDuckGoAPI)
	if err != nil {
		chDuckDuckGo <- err.Error()
		return
	}
	var duckDuckGoResponse DuckDuckResponse
	if err := json.Unmarshal(bodyDuck, &duckDuckGoResponse); err != nil {
		chDuckDuckGo <- err.Error()
		return
	}

	chDuckDuckGo <- duckDuckGoResponse.RelatedTopics[0].Text

}

func googleSearch(chGoogle chan string, q string, googleAPI *string) {
	*googleAPI = fmt.Sprint(*googleAPI + "&q=" + q)
	_, body, err := fasthttp.Get(nil, *googleAPI)
	if err != nil {
		chGoogle <- err.Error()
		return
	}
	var googleResponse GoogleResponse

	if err := json.Unmarshal(body, &googleResponse); err != nil {
		chGoogle <- err.Error()
		return
	}

	chGoogle <- googleResponse.Items[0].Snippet

}

func twitterTweet(chTwitter chan string, q string, twitterAPI *string) {

	*twitterAPI = fmt.Sprint(*twitterAPI + "?q=" + q)

	c := oauth.NewConsumer(
		consumerKey,
		consumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		})

	client, err := c.MakeHttpClient(&oauth.AccessToken{
		Token:  accessToken,
		Secret: accessTokenSecret,
	})
	if err != nil {
		chTwitter <- err.Error()
		return
	}

	response, err := client.Get(
		*twitterAPI)
	if err != nil {
		chTwitter <- err.Error()
		return
	}
	twitterBody, _ := ioutil.ReadAll(response.Body)

	var twitterResponse twitterResponse
	if err := json.Unmarshal(twitterBody, &twitterResponse); err != nil {
		chTwitter <- err.Error()
		return
	}

	chTwitter <- twitterResponse.Statuses[0].Text

}
