package main

type FeedEntriesSortByPubDate []*FeedEntry

func (fe FeedEntriesSortByPubDate) Len() int           { return len(fe) }
func (fe FeedEntriesSortByPubDate) Swap(i, j int)      { fe[i], fe[j] = fe[j], fe[i] }
func (fe FeedEntriesSortByPubDate) Less(i, j int) bool { return !fe[i].PubDate.After(*fe[j].PubDate) }
