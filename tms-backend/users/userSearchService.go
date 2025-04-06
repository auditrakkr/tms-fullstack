package users

import "github.com/elastic/go-elasticsearch/v8"

type UserSearchService struct {
	client *elasticsearch.Client
	
}