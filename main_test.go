package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
		// пока сравнивать не будем, а просто выведем ответы
		// удалите потом этот вывод
		fmt.Println(response.Body.String())
	}

}

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requestsBad := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, a := range requestsBad {
		responseBad := httptest.NewRecorder()
		reqBad := httptest.NewRequest("GET", a.request, nil)

		handler.ServeHTTP(responseBad, reqBad)
		assert.Equal(t, a.status, responseBad.Code)
		assert.Equal(t, a.message, strings.TrimSpace(responseBad.Body.String()))
		fmt.Println(responseBad.Body.String())
	}
}

func TestCafeCount(p *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requestsCount := []struct {
		request string
		city    string
		count   int // передаваемое значение count
		want    int // ожидаемое количество кафе в ответе
	}{
		{"/cafe?city=moscow&count=0", "moscow", 0, 0},
		{"/cafe?city=moscow&count=1", "moscow", 1, 1},
		{"/cafe?city=moscow&count=2", "moscow", 2, 2},
		{"/cafe?city=moscow&count=100", "moscow", 100, min(100, len(cafeList["moscow"]))}, // Ожидаем не более 100
	}

	for _, tt := range requestsCount {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", tt.request, nil)

		handler.ServeHTTP(response, req)
		body := strings.TrimSpace(response.Body.String())

		//Проверка количества кафе
		var actualCount int
		if body != "" {
			cafes := strings.Split(body, ",")
			actualCount = len(cafes)
		}
		assert.Equal(p, tt.want, actualCount)
		// пока сравнивать не будем, а просто выведем ответы
		// удалите потом этот вывод
		fmt.Println(response.Body.String())
	}
}

func TestCafeSearch(s *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		name      string
		search    string
		wantCount int
	}{
		{"Search 'фасоль'", "фасоль", 0},
		{"Search 'кофе'", "кофе", 2},
		{"Search 'вилка'", "вилка", 1},
	}

	for _, tt := range tests {
		s.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/cafe?city=moscow&search=%s", tt.search)
			req := httptest.NewRequest("GET", url, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)

			assert.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			// Проверка количества найденных кафе
			assert.Equal(t, tt.wantCount, len(cafes),
				"Для search='%s' ожидалось %d кафе, получили %d",
				tt.search, tt.wantCount, len(cafes))

			// Проверка что каждое кафе содержит искомую подстроку (без учёта регистра)
			searchLower := strings.ToLower(tt.search)
			for _, cafe := range cafes {
				cafeLower := strings.ToLower(strings.TrimSpace(cafe))
				assert.True(t, strings.Contains(cafeLower, searchLower),
					"Кафе '%s' не содержит подстроку '%s'", cafe, tt.search)
			}
			fmt.Println(response.Body.String())
		})

	}
}
