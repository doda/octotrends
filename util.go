package main

// Some repos are attributed to languages they have little to do with
var BLOCKEDREPOS = []string{
	"996icu/996.ICU",
	"public-apis/public-apis",
	"CyC2018/CS-Notes",
	"awesome-selfhosted/awesome-selfhosted",
	"jaywcjlove/awesome-mac",
	"bayandin/awesome-awesomeness",
	"donnemartin/system-design-primer",
}

func StringValue(sp *string) string {
	if sp == nil {
		return ""
	}
	return *sp
}

func IntValue(ip *int) int {
	if ip == nil {
		return 0
	}
	return *ip
}

func RepoLangBlocked(name string) bool {
	for _, blocked := range BLOCKEDREPOS {
		if blocked == name {
			return true
		}
	}
	return false
}
