package sse

import (
	"fmt"
	"log"
	"net/http"
)

// Broker управляет подключениями SSE
type Broker struct {
	// Канал для регистрации новых клиентов
	newClients chan chan []byte
	// Канал для удаления отключившихся клиентов
	closingClients chan chan []byte
	// Канал для получения сообщений на рассылку
	Notifier chan []byte
	// Карта активных клиентских каналов
	clients map[chan []byte]bool
}

// NewBroker инициализирует брокера
func NewBroker() *Broker {
	return &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}
}

// Start запускает бесконечный цикл обработки событий в отдельной горутине.
// Это гарантирует потокобезопасность map (без использования мьютексов).
func (b *Broker) Start() {
	for {
		select {
		case s := <-b.newClients:
			b.clients[s] = true
			log.Printf("Client added. Total clients: %d", len(b.clients))

		case s := <-b.closingClients:
			delete(b.clients, s)
			log.Printf("Removed client. Total clients: %d", len(b.clients))

		case event := <-b.Notifier:
			for clientMessageChan := range b.clients {
				select {
				case clientMessageChan <- event:
				default:
					// Если канал клиента переполнен или заблокирован, удаляем его
					delete(b.clients, clientMessageChan)
					close(clientMessageChan)
				}
			}
		}
	}
}

// ServeHTTP обрабатывает запросы клиентов и держит соединение открытым
func (b *Broker) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Проверяем поддержку Flush (обязательно для SSE)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовки для SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Важно для обхода CORS, если фронт и бэк на разных портах локально
	w.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan []byte)
	b.newClients <- messageChan

	// Отписываем клиента при разрыве соединения
	defer func() {
		b.closingClients <- messageChan
		close(messageChan)
	}()

	ctx := req.Context()

	for {
		select {
		case msg := <-messageChan:
			// Формат SSE требует префикса "data: " и завершения двумя переносами строк
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ctx.Done():
			// Клиент закрыл вкладку
			return
		}
	}
}