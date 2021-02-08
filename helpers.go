package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// exists returns whether the given file or directory exists or not
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func fatal(err error) {
	if err != nil {
		//todo: time of the error
		errorLogFile.WriteString(fmt.Sprintln(err))
	}
}

func closeApp(err error) {
	panic(err)
}

func flowPrintln(text string) {
	for _, v := range text {
		time.Sleep(4 * time.Millisecond)
		fmt.Print(string(v))
	}
	fmt.Println("")
}

func Nl2br(str string) string {
	str = strings.Replace(str, "\r\n", "<br>", -1)
	str = strings.Replace(str, "\n\r", "<br>", -1)
	str = strings.Replace(str, "\n", "<br>", -1)
	str = strings.Replace(str, "\r", "<br>", -1)
	return str
}

func getCookieByName(cookie []*http.Cookie, name string) string {
	cookieLen := len(cookie)
	result := ""
	for i := 0; i < cookieLen; i++ {
		if cookie[i].Name == name {
			result = cookie[i].Value
		}
	}
	return result
}

func removeUserStateFromSlice(s []UserState, i int) []UserState {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}