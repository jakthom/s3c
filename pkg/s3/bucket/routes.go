package s3bucket

import "github.com/gorilla/mux"

const (
	Route              = `/{bucket:[a-zA-Z0-9\-_\.]{1,255}}`
	trailingSlashRoute = `/{bucket:[a-zA-Z0-9\-_\.]{1,255}}/`
)

func AddSubrouter(router *mux.Router, handler *BucketHandler) error {
	subrouter := router.PathPrefix(Route).Subrouter()
	attachRoutes(subrouter, handler)
	trailingSlashSubrouter := router.PathPrefix(trailingSlashRoute).Subrouter()
	attachRoutes(trailingSlashSubrouter, handler)
	return nil
}

// AttachRoutes attaches the routes for the bucket handler to the router
// func attachBucketRoutes(router *mux.Router, handler *BucketHandler, multipartHandler *multipartHandler, objectHandler *objectHandler) {
func attachRoutes(router *mux.Router, handler *BucketHandler) {
	// router.Methods("GET").Queries("versioning", "").HandlerFunc(handler.versioning) // TODO
	// router.Methods("PUT").Queries("versioning", "").HandlerFunc(handler.setVersioning) // TODO
	// router.Methods("GET").Queries("versions", "").HandlerFunc(handler.listVersions) // TODO
	// router.Methods("GET").Queries("uploads", "").HandlerFunc(multipartHandler.list) // TODO
	router.Methods("GET").Queries("location", "").HandlerFunc(handler.Location)
	router.Methods("GET", "HEAD").HandlerFunc(handler.Get)
	router.Methods("PUT").HandlerFunc(handler.Put)
	// router.Methods("POST").Queries("delete", "").HandlerFunc(objectHandler.post)
	router.Methods("DELETE").HandlerFunc(handler.Del)
}
