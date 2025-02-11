package s3handler

import "github.com/gorilla/mux"

// AddNotImplementedRoutes adds routes to the provided router to return
// NotImplemented xml errs.
func AddNotImplementedRoutes(router *mux.Router) {
	//
	router.Methods("GET", "PUT").Queries("accelerate", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT").Queries("acl", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("analytics", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("cors", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("encryption", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("inventory", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("lifecycle", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT").Queries("logging", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("metrics", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT").Queries("notification", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT").Queries("object-lock", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("policy", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET").Queries("policyStatus", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("publicAccessBlock", "").HandlerFunc(NotImplementedHandler())
	router.Methods("PUT", "DELETE").Queries("replication", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT").Queries("requestPayment", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("tagging", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("website", "").HandlerFunc(NotImplementedHandler())
	//
	router.Methods("GET", "PUT").Queries("acl", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT").Queries("legal-hold", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT").Queries("retention", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET", "PUT", "DELETE").Queries("tagging", "").HandlerFunc(NotImplementedHandler())
	router.Methods("GET").Queries("torrent", "").HandlerFunc(NotImplementedHandler())
	router.Methods("POST").Queries("restore", "").HandlerFunc(NotImplementedHandler())
	router.Methods("POST").Queries("select", "").HandlerFunc(NotImplementedHandler())
	// catch-all for POST calls that aren't using the delete subresource
	router.Methods("POST").HandlerFunc(NotImplementedHandler())
}
