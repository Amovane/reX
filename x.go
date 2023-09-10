package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	urlutil "net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	SCRAPER "github.com/n0madic/twitter-scraper"
)

type X struct {
	uname   string
	upwd    string
	scraper *SCRAPER.Scraper
}

func New(uname string, upwd string) X {
	return X{
		uname:   uname,
		upwd:    upwd,
		scraper: SCRAPER.New(),
	}
}

func (x *X) login() error {
	return x.scraper.Login(x.uname, x.upwd)
}

func (x *X) IsLoggedIn() bool {
	return x.scraper.IsLoggedIn()
}

func (x *X) GetFollowingsByScreenName(user string, cursor *string) (resp []Legacy, nextCursor *string) {
	uid, _ := x.scraper.GetUserIDByScreenName(user)
	return x.GetFollowingsById(uid, cursor)
}

func (x *X) GetFollowingsById(uid string, cursor *string) (resp []Legacy, nextCursor *string) {
	var csrfToken string
	cookies := Map(x.scraper.GetCookies(), func(field *http.Cookie) string {
		if field.Name == "ct0" {
			csrfToken = field.Value
		}
		return field.String()
	})
	cookiesStr := strings.Join(cookies, ";")
	variables := T{
		"userId":                 uid,
		"count":                  20,
		"includePromotedContent": false,
	}
	if cursor != nil {
		variables["cursor"] = *cursor
	}
	features := T{
		"rweb_lists_timeline_redesign_enabled":                                    true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                false,
		"tweet_awards_web_tipping_enabled":                                        false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_media_download_video_enabled":                             false,
		"responsive_web_enhance_cards_enabled":                                    false,
	}
	variablesJson, _ := json.Marshal(variables)
	featuresJson, _ := json.Marshal(features)
	query := fmt.Sprintf(`variables=%s&features=%s`, variablesJson, featuresJson)
	values, _ := urlutil.ParseQuery(query)
	url := fmt.Sprintf(`https://twitter.com/i/api/graphql/%s?%s`, Following.Path(), values.Encode())

	var response Response
	var err error
	client := resty.New()
	client.
		R().
		SetHeaders(
			StringMap{
				"authority":                 "twitter.com",
				"accept":                    "*/*",
				"accept-language":           "zh-CN,zh;q=0.9,en;q=0.8",
				"authorization":             "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
				"content-type":              "application/json",
				"cookie":                    cookiesStr,
				"sec-ch-ua":                 `"Chromium";v="116", "Not)A;Brand";v="24", "Google Chrome";v="116"`,
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        `"macOS"`,
				"sec-fetch-dest":            "empty",
				"sec-fetch-mode":            "cors",
				"sec-fetch-site":            "same-origin",
				"user-agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
				"x-client-transaction-id":   "b8qgMoxBUsxfsTLOsISkGvJ9/Atx8/2g/teNmizHQONJLNnUNCKMBxHt2eRlE6jkzOuM9G/tZ/mwb0KT9PAok5bnqm6vbg",
				"x-client-uuid":             "cc8bdbff-b377-4ffd-b53e-5767f6e50ba4",
				"x-csrf-token":              csrfToken,
				"x-twitter-active-user":     "yes",
				"x-twitter-auth-type":       "OAuth2Session",
				"x-twitter-client-language": "en",
			},
		).
		SetResult(&response).
		SetError(&err).
		Get(url)

	instructions := response.Data.User.Result.Timeline.Timeline.Instructions
	resp = make([]Legacy, 0)
	for _, i := range instructions {
		for _, e := range i.Entries {
			cursorType := e.Content.CursorType
			if cursorType != nil && *cursorType == "Bottom" {
				cursor = e.Content.Value
			}
			item := e.Content.ItemContent
			if item == nil {
				continue
			}
			resp = append(resp, item.UserResults.Result.Legacy)
		}
	}
	return resp, cursor
}