package main

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
