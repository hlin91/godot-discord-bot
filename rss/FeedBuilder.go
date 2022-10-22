package rss

import "golang.org/x/net/html"

type FeedBuilder struct {
	ItemProvider                *ItemProvider
	NumImages                   int
	ImageNodeFilterStrategy     func(*html.Node) bool
	LogoImageNodeFilterStrategy func(*html.Node) bool
	ImageLinkExtractionStrategy func(*html.Node) string
	ImageLinkTransformStrategy  func(string) string
	GetChannelIdStrategy        func() string
}

func NewFeedBuilder() *FeedBuilder {
	return &FeedBuilder{
		ItemProvider:                nil,
		NumImages:                   1,
		ImageNodeFilterStrategy:     DefaultFilterStrategy(),
		LogoImageNodeFilterStrategy: DefaultFilterStrategy(),
		ImageLinkExtractionStrategy: DefaultExtractionStrategy(),
		ImageLinkTransformStrategy:  DefaultTransformStrategy(),
		GetChannelIdStrategy: func() string {
			return ""
		},
	}
}

func (f *FeedBuilder) WithItemProvider(itemProvider *ItemProvider) *FeedBuilder {
	f.ItemProvider = itemProvider
	return f
}

func (f *FeedBuilder) WithImageLimit(limit int) *FeedBuilder {
	f.NumImages = limit
	return f
}

func (f *FeedBuilder) WithImageNodeFilterStrategy(strategy func(*html.Node) bool) *FeedBuilder {
	f.ImageNodeFilterStrategy = strategy
	return f
}

func (f *FeedBuilder) WithLogoImageNodeFilterStrategy(strategy func(*html.Node) bool) *FeedBuilder {
	f.LogoImageNodeFilterStrategy = strategy
	return f
}

func (f *FeedBuilder) WithImageLinkExtractionStrategy(strategy func(*html.Node) string) *FeedBuilder {
	f.ImageLinkExtractionStrategy = strategy
	return f
}

func (f *FeedBuilder) WithImageLinkTransformStrategy(strategy func(string) string) *FeedBuilder {
	f.ImageLinkTransformStrategy = strategy
	return f
}

func (f *FeedBuilder) WithGetChannelIdStrategy(strategy func() string) *FeedBuilder {
	f.GetChannelIdStrategy = strategy
	return f
}

func (f *FeedBuilder) Build() *Feed {
	return &Feed{
		ItemProvider:                f.ItemProvider,
		NumImages:                   f.NumImages,
		ImageNodeFilterStrategy:     &f.ImageNodeFilterStrategy,
		LogoImageNodeFilterStrategy: &f.LogoImageNodeFilterStrategy,
		ImageLinkExtractionStrategy: &f.ImageLinkExtractionStrategy,
		ImageLinkTransformStrategy:  &f.ImageLinkTransformStrategy,
		GetChannelIdStrategy:        &f.GetChannelIdStrategy,
	}
}
