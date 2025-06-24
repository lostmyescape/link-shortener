package save

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/lostmyescape/url-shortener/internal/http-server/handlers/url/save/mocks"
	"github.com/lostmyescape/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/lostmyescape/url-shortener/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
		wantCode  int
	}{
		{
			name:     "Success",
			alias:    "google",
			url:      "https://google.com",
			wantCode: http.StatusOK,
		},
		{
			name:     "Empty alias",
			alias:    "",
			url:      "https://google.com",
			wantCode: http.StatusOK,
		},
		{
			name:      "Empty URL",
			url:       "",
			alias:     "example",
			respError: "field URL is a required field",
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "Invalid URL",
			url:       "some invalid URL",
			alias:     "some_alias",
			respError: "field URL is not a valid URL",
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "SaveURL Error",
			alias:     "test_alias",
			url:       "https://google.com",
			respError: "failed to add URL",
			mockError: errors.New("unexpected error"),
			wantCode:  http.StatusInternalServerError,
		},
		{
			name:      "URL already exists",
			url:       "https://google.com",
			alias:     "go",
			wantCode:  http.StatusConflict,
			respError: "URL already exists",
			mockError: storage.ErrURLExists,
		},
	}

	for _, tc := range cases {
		// создание локальной копии переменной для работы в параллельных тестах
		tc := tc // (захват переменной)

		// запуск подтеста с именем tc.name
		t.Run(tc.name, func(t *testing.T) {
			// позволяет тесту выполняться параллельно с другими
			t.Parallel()

			// настройка мока:
			// urlSaverMock имитирует реальное поведение URLSaver
			urlSaverMock := mocks.NewURLSaver(t)

			// мок настраиваться только если:
			// ожидается успешный ответ или задана ошибка для мока
			if tc.respError == "" || tc.mockError != nil {
				// мок ожидать вызова SaveURL с аргументами tc.url и любым string
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError). // возвращает 1 и ошибку
					Once()                          // метод вызывается только один раз
			}

			// создание хендлера: принимает заглушку и мок
			handler := New(slogdiscard.NewDiscardLogger(), urlSaverMock)

			// тело запроса в JSON
			bodyBytes, err := json.Marshal(map[string]string{
				"url":   tc.url,
				"alias": tc.alias,
			})
			require.NoError(t, err)

			// создает POST запрос к /save
			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader(bodyBytes))
			require.NoError(t, err)

			// Запись ответа:
			// 1. запись ответа сервера
			rr := httptest.NewRecorder()
			// 2. выполняет запрос
			handler.ServeHTTP(rr, req)

			// сравнение статусов ответов
			require.Equal(t, tc.wantCode, rr.Code)

			body := rr.Body.String()

			var resp Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			// смотрим что ошибка, которую вернул хендлер == ошибке которая определена в тест кейсе
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
