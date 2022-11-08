package main

type PocketConfig struct {
	ConsumerKey string `split_words:"true" required:"true"`
	AccessToken string `split_words:"true" required:"true"`
}

type Config struct {
	Title string `split_words:"true" default:"Pocket feed"`
	URL   string `split_words:"true" default:"https://getpocket.com"`
}
