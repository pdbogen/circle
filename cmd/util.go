package main

import "strconv"

func human(bytes int) string {
	if bytes < 1024 {
		return strconv.Itoa(bytes) + "B"
	}
	if bytes < 1024*1024 {
		return strconv.Itoa(bytes/1024) + "KiB"
	}
	if bytes < 1024*1024*1024 {
		return strconv.Itoa(bytes/1024/1024) + "MiB"
	}
	return strconv.Itoa(bytes/1024/1024/1024) + "GiB"
}
