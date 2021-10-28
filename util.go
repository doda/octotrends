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

func RepoLangDoesntCount(name string) bool {
	for _, blocked := range BLOCKEDREPOS {
		if blocked == name {
			return true
		}
	}
	return false
}

// Safely deref a string-pointer
func StringValue(sp *string) string {
	if sp == nil {
		return ""
	}
	return *sp
}

// Safely deref an int-pointer
func IntValue(ip *int) int {
	if ip == nil {
		return 0
	}
	return *ip
}
