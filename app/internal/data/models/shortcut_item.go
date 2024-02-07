package models

type ShortcutItem struct {
	Id  string `dynamodbav:"id"`
	Url string `dynamodbav:"url"`
}
