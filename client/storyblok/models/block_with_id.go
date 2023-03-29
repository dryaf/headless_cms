package models

import "time"

type SimpleBlockskWithID struct {
	Story struct {
		Name        string    `json:"name"`
		CreatedAt   time.Time `json:"created_at"`
		PublishedAt time.Time `json:"published_at"`
		ID          int       `json:"id"`
		UUID        string    `json:"uuid"`
		Content     struct {
			UID       string           `json:"_uid"`
			Body      []map[string]any `json:"body"`
			Component string           `json:"component"`
		} `json:"content"`
		Slug             string        `json:"slug"`
		FullSlug         string        `json:"full_slug"`
		SortByDate       interface{}   `json:"sort_by_date"`
		Position         int           `json:"position"`
		TagList          []interface{} `json:"tag_list"`
		IsStartpage      bool          `json:"is_startpage"`
		ParentID         interface{}   `json:"parent_id"`
		MetaData         interface{}   `json:"meta_data"`
		GroupID          string        `json:"group_id"`
		FirstPublishedAt time.Time     `json:"first_published_at"`
		ReleaseID        interface{}   `json:"release_id"`
		Lang             string        `json:"lang"`
		Path             interface{}   `json:"path"`
		Alternates       []interface{} `json:"alternates"`
		DefaultFullSlug  interface{}   `json:"default_full_slug"`
		TranslatedSlugs  interface{}   `json:"translated_slugs"`
	} `json:"story"`
	Cv    int           `json:"cv"`
	Rels  []interface{} `json:"rels"`
	Links []interface{} `json:"links"`
}
