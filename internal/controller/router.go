package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"gitlab.com/dgb9/todo-api/internal/data"
)

func StartRouter(db *sql.DB, config data.ServerConfig) error {
	context := config.Context
	mux := http.NewServeMux()
	mux.Handle(fmt.Sprintf("GET %s/", context), rootController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/login", context), loginController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/addLogin", context), addLoginController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/removeLogin", context), removeLoginController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/updatePriority", context), updatePriorityController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/toggleCompleted", context), toggleCompletedController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/switchSeqno", context), switchSeqNoController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/logout", context), logoutController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/removeTodo", context), removeTodoController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/retrieveTodo", context), retrieveTodoController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/updateDue", context), updateDueController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/updateTodo", context), updateTodoController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/removeAttachment", context), removeAttachmentController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/search", context), searchTodoController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/multipleTodo", context), multipleTodoController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/bulkUpdate", context), bulkUpdateController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/removeItem", context), removeItemController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/updateItem", context), updateItemController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/updateAttachment", context), updateAttachmentController(db, config))
	mux.Handle(fmt.Sprintf("GET %s/flaggedItems", context), flaggedItemsController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/searchItem", context), searchItemController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/addBulkAttachment", context), addBulkAttachmentController(db, config))
	mux.Handle(fmt.Sprintf("GET %s/processDownload", context), downloadController(db, config))
	mux.Handle(fmt.Sprintf("POST %s/password", context), passwordPostController(db, config))
	mux.Handle(fmt.Sprintf("PUT %s/password", context), passwordPutController(db, config))

	corsRouter := corsMiddleware(mux)
	loggedRouter := loggingMiddleward(corsRouter)

	return http.ListenAndServe(config.Address, loggedRouter)
}

func loggingMiddleward(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r) // Call the next handler
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins (adjust in prod)
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Allow credentials if needed
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Allow headers
		w.Header().Set("Access-Control-Allow-Headers", "*")

		// Allow methods
		w.Header().Set("Access-Control-Allow-Methods", "*")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Continue to actual handler
		next.ServeHTTP(w, r)
	})
}
