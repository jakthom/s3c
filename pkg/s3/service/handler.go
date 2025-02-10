package s3service

import (
	"net/http"
	"time"

	s3util "github.com/jakthom/s3c/pkg/s3/util"
)

type ServiceHandler struct {
	Controller ServiceController
}

func (h *ServiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	result, err := h.Controller.ListBuckets(r)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	// some clients (e.g. minio-python) can't handle sub-seconds in datetime
	// output
	for _, bucket := range result.Buckets {
		bucket.CreationDate = bucket.CreationDate.UTC().Round(time.Second)
	}

	s3util.WriteXML(w, r, http.StatusOK, result)
}
