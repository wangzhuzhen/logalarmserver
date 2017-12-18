package types

import (
	"time"
	"net/http"
)

type Rule struct {
	RuleName      string   `json:"rulename"`
	KeyWord        string   `json:"keyword"`
	KeywordIndex        int   `json:"keywordindex"`
	TimeWindow 	       int   `json:"timewindow"`
	ThresholdNum    int   `json:"thresholdnum"`
	Due       time.Time  `json:"due"`
}

type Rules []Rule


type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route