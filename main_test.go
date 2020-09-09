package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncrementValue(t *testing.T) {
	testCases := []struct {
		key      string
		value    string
		expected string
	}{
		{
			key:      "age",
			value:    "-10",
			expected: "-10",
		},
		{
			key:      "age",
			value:    "12",
			expected: "2",
		},
		{
			key:      "age",
			value:    "22",
			expected: "24",
		},
	}
	handler := http.HandlerFunc(incrementValue)
	for _, testCase := range testCases {
		t.Run(testCase.value, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "localhost:8081/redis/incr?key="+testCase.key+"&value="+testCase.value, nil)
			handler.ServeHTTP(rec, req)
			result := rec.Result()
			defer result.Body.Close()
			resultBytes, _ := ioutil.ReadAll(result.Body)
			actualValueString := strings.Split(string(resultBytes), ":")[1]
			actualValueString = strings.Split(string(actualValueString), "}")[0]
			assert.Equal(t, testCase.expected, actualValueString)
		})
	}
}

func TestGetSignature(t *testing.T) {
	testCases := []struct {
		text     string
		key      string
		expected string
	}{
		{
			text:     "text",
			key:      "key2",
			expected: "\"ffe0039bdc9a19c83cd980779c276bddb761290b137f60b871749cee4c71d2f0\"",
		},
		{
			text:     "tex",
			key:      "key2",
			expected: "\"Invalid text\"",
		},
		{
			text:     "text",
			key:      "key",
			expected: "\"Invalid key\"",
		},
		{
			text:     "sometext",
			key:      "somekey",
			expected: "\"1f8314daed7b01d17cd907a218e7ed057860c03b41d1255db260517d35d91fa7\"",
		},
	}
	handler := http.HandlerFunc(getSignature)
	for _, testCase := range testCases {
		t.Run(testCase.text, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "localhost:8081/sign/hmacsha512?text="+testCase.text+"&key="+testCase.key, nil)
			handler.ServeHTTP(rec, req)
			result := rec.Result()
			defer result.Body.Close()
			resultBytes, _ := ioutil.ReadAll(result.Body)
			actualValueString := strings.Split(string(resultBytes), ":")[1]
			actualValueString = strings.Split(string(actualValueString), "}")[0]
			assert.Equal(t, testCase.expected, actualValueString)
		})
	}
}

func TestInsertUser(t *testing.T) {
	testCases := []struct {
		name     string
		age      string
		expected string
	}{
		{
			name:     "Oleg",
			age:      "12",
			expected: "1",
		},
		{
			name:     "Ivan",
			age:      "34",
			expected: "2",
		},
		{
			name:     "Semen",
			age:      "16",
			expected: "3",
		},
		{
			name:     "Viktor",
			age:      "34",
			expected: "4",
		},
	}
	handler := http.HandlerFunc(insertUser)
	for _, testCase := range testCases {
		t.Run(testCase.name+" "+testCase.age, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "localhost:8081/postgres/users?name="+testCase.name+"&age="+testCase.age, nil)
			handler.ServeHTTP(rec, req)
			result := rec.Result()
			defer result.Body.Close()
			resultBytes, _ := ioutil.ReadAll(result.Body)
			actualValueString := strings.Split(string(resultBytes), ":")[1]
			actualValueString = strings.Split(string(actualValueString), "}")[0]
			assert.Equal(t, testCase.expected, actualValueString)
		})
	}
}
