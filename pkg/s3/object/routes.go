package s3object

import "github.com/gorilla/mux"

const (
	Route = `/{bucket:[a-zA-Z0-9\-_\.]{1,255}}/{key:.+}`
)

func AddSubrouter(router *mux.Router, handler *ObjectHandler) error {
	subrouter := router.PathPrefix(Route).Subrouter()
	attachRoutes(subrouter, handler)
	return nil
}

func attachRoutes(router *mux.Router, handler *ObjectHandler) {
	// router.Methods("GET").Queries("uploadId", "").HandlerFunc(multipartHandler.listChunks)
	// router.Methods("POST").Queries("uploads", "").HandlerFunc(multipartHandler.init)
	// router.Methods("POST").Queries("uploadId", "").HandlerFunc(multipartHandler.complete)
	// router.Methods("PUT").Queries("uploadId", "").HandlerFunc(multipartHandler.put)
	// router.Methods("DELETE").Queries("uploadId", "").HandlerFunc(multipartHandler.del)
	// router.Methods("HEAD").HandlerFunc(handler.Head)
	router.Methods("GET", "HEAD").HandlerFunc(handler.Get)
	router.Methods("PUT").Headers("x-amz-copy-source", "").HandlerFunc(handler.Copy)
	router.Methods("PUT").HandlerFunc(handler.Put)
	router.Methods("DELETE").HandlerFunc(handler.Del)
}
