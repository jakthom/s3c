package s3object

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	s3error "github.com/jakthom/s3c/pkg/s3/error"
	s3util "github.com/jakthom/s3c/pkg/s3/util"
)

type objectHandler struct {
	controller ObjectController
}

func (h *objectHandler) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	versionId := r.FormValue("versionId")

	result, err := h.controller.GetObject(r, bucket, key, versionId)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	if result.ETag != "" {
		w.Header().Set("ETag", s3util.AddETagQuotes(result.ETag))
	}
	if result.Version != "" {
		w.Header().Set("x-amz-version-id", result.Version)
	}

	if result.DeleteMarker {
		w.Header().Set("x-amz-delete-marker", "true")
		s3util.WriteError(w, r, s3error.NoSuchKeyError(r))
		return
	}

	http.ServeContent(w, r, key, result.ModTime, result.Content)
}

func (h *objectHandler) copy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	destBucket := vars["bucket"]
	destKey := vars["key"]

	var srcBucket string
	var srcKey string
	srcURL, err := url.Parse(r.Header.Get("x-amz-copy-source"))
	if err != nil {
		s3util.WriteError(w, r, s3error.InvalidArgumentError(r))
		return
	}
	srcPath := strings.SplitN(srcURL.Path, "/", 3)
	if len(srcPath) == 2 {
		srcBucket = srcPath[0]
		srcKey = srcPath[1]
	} else if len(srcPath) == 3 {
		if srcPath[0] != "" {
			s3util.WriteError(w, r, s3error.InvalidArgumentError(r))
			return
		}
		srcBucket = srcPath[1]
		srcKey = srcPath[2]
	} else {
		s3util.WriteError(w, r, s3error.InvalidArgumentError(r))
		return
	}
	srcVersionID := srcURL.Query().Get("versionId")

	if srcBucket == "" {
		s3util.WriteError(w, r, s3error.InvalidBucketNameError(r))
		return
	}
	if srcKey == "" {
		s3util.WriteError(w, r, s3error.NoSuchKeyError(r))
		return
	}
	if srcBucket == destBucket && srcKey == destKey && srcVersionID == "" {
		// If we ever add support for object metadata, this error should not
		// trigger in the case where metadata is changed, since it is a valid
		// way to alter the metadata of an object
		s3util.WriteError(w, r, s3error.InvalidRequestError(r, "source and destination are the same"))
		return
	}

	ifMatch := r.Header.Get("x-amz-copy-source-if-match")
	ifNoneMatch := r.Header.Get("x-amz-copy-source-if-none-match")
	ifUnmodifiedSince := r.Header.Get("x-amz-copy-source-if-unmodified-since")
	ifModifiedSince := r.Header.Get("x-amz-copy-source-if-modified-since")

	getResult, err := h.controller.GetObject(r, srcBucket, srcKey, srcVersionID)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}
	if getResult.DeleteMarker {
		s3util.WriteError(w, r, s3error.NoSuchKeyError(r))
		return
	}

	if !s3util.CheckIfMatch(ifMatch, getResult.ETag) {
		s3util.WriteError(w, r, s3error.PreconditionFailedError(r))
		return
	}

	if !s3util.CheckIfNoneMatch(ifNoneMatch, getResult.ETag) {
		s3util.WriteError(w, r, s3error.PreconditionFailedError(r))
		return
	}

	if !s3util.CheckIfUnmodifiedSince(ifUnmodifiedSince, getResult.ModTime) {
		s3util.WriteError(w, r, s3error.PreconditionFailedError(r))
		return
	}

	if !s3util.CheckIfModifiedSince(ifModifiedSince, getResult.ModTime) {
		s3util.WriteError(w, r, s3error.PreconditionFailedError(r))
		return
	}

	destVersionID, err := h.controller.CopyObject(r, srcBucket, srcKey, getResult, destBucket, destKey)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	if getResult.Version != "" {
		w.Header().Set("x-amz-copy-source-version-id", getResult.Version)
	}

	if destVersionID != "" {
		w.Header().Set("x-amz-version-id", srcVersionID)
	}

	marshallable := struct {
		XMLName      xml.Name  `xml:"http://s3.amazonaws.com/doc/2006-03-01/ CopyObjectResult"`
		LastModified time.Time `xml:"LastModified"`
		ETag         string    `xml:"ETag"`
	}{
		LastModified: getResult.ModTime,
		ETag:         getResult.ETag,
	}

	s3util.WriteXML(w, r, http.StatusOK, marshallable)
}

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	transferEncoding := r.Header["Transfer-Encoding"]
	identity := false
	for _, headerValue := range transferEncoding {
		if headerValue == "identity" {
			identity = true
		}
	}
	if len(transferEncoding) == 0 || identity {
		if err := s3util.RequireContentLength(r); err != nil {
			s3util.WriteError(w, r, err)
			return
		}
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	chunked := r.Header.Get("X-Amz-Content-Sha256") == "STREAMING-AWS4-HMAC-SHA256-PAYLOAD"

	var body io.ReadCloser
	if chunked {
		signingKey := []byte(vars["authSignatureKey"])
		seedSignature := vars["authSignature"]
		timestamp := vars["authSignatureTimestamp"]
		date := vars["authSignatureDate"]
		region := vars["authSignatureRegion"]
		body = s3util.NewChunkedReader(r.Body, signingKey, seedSignature, timestamp, date, region)
	} else {
		body = r.Body
	}

	result, err := h.controller.PutObject(r, bucket, key, body)
	if err != nil {
		if err == s3util.InvalidChunk {
			s3util.WriteError(w, r, s3error.SignatureDoesNotMatchError(r))
		} else {
			s3util.WriteError(w, r, err)
		}
		return
	}

	if result.ETag != "" {
		w.Header().Set("ETag", s3util.AddETagQuotes(result.ETag))
	}
	if result.Version != "" {
		w.Header().Set("x-amz-version-id", result.Version)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	versionId := r.FormValue("versionId")

	result, err := h.controller.DeleteObject(r, bucket, key, versionId)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	if result.Version != "" {
		w.Header().Set("x-amz-version-id", result.Version)
	}
	if result.DeleteMarker {
		w.Header().Set("x-amz-delete-marker", "true")
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *objectHandler) post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	payload := struct {
		XMLName xml.Name `xml:"Delete"`
		Quiet   bool     `xml:"Quiet"`
		Objects []struct {
			Key     string `xml:"Key"`
			Version string `xml:"VersionId"`
		} `xml:"Object"`
	}{}
	if err := s3util.ReadXMLBody(r, &payload); err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	marshallable := struct {
		XMLName xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ DeleteResult"`
		Deleted []struct {
			Key                 string `xml:"Key"`
			Version             string `xml:"Version,omitempty"`
			DeleteMarker        bool   `xml:"Code,omitempty"`
			DeleteMarkerVersion string `xml:"DeleteMarkerVersionId,omitempty"`
		} `xml:"Deleted"`
		Errors []struct {
			Key     string `xml:"Key"`
			Code    string `xml:"Code"`
			Message string `xml:"Message"`
		} `xml:"Error"`
	}{
		Deleted: []struct {
			Key                 string `xml:"Key"`
			Version             string `xml:"Version,omitempty"`
			DeleteMarker        bool   `xml:"Code,omitempty"`
			DeleteMarkerVersion string `xml:"DeleteMarkerVersionId,omitempty"`
		}{},
		Errors: []struct {
			Key     string `xml:"Key"`
			Code    string `xml:"Code"`
			Message string `xml:"Message"`
		}{},
	}

	for _, object := range payload.Objects {
		result, err := h.controller.DeleteObject(r, bucket, object.Key, object.Version)
		if err != nil {
			s3Err := s3error.NewGenericError(r, err)

			marshallable.Errors = append(marshallable.Errors, struct {
				Key     string `xml:"Key"`
				Code    string `xml:"Code"`
				Message string `xml:"Message"`
			}{
				Key:     object.Key,
				Code:    s3Err.Code,
				Message: s3Err.Message,
			})
		} else {
			deleteMarkerVersion := ""
			if result.DeleteMarker {
				deleteMarkerVersion = result.Version
			}

			if !payload.Quiet {
				marshallable.Deleted = append(marshallable.Deleted, struct {
					Key                 string `xml:"Key"`
					Version             string `xml:"Version,omitempty"`
					DeleteMarker        bool   `xml:"Code,omitempty"`
					DeleteMarkerVersion string `xml:"DeleteMarkerVersionId,omitempty"`
				}{
					Key:                 object.Key,
					Version:             object.Version,
					DeleteMarker:        result.DeleteMarker,
					DeleteMarkerVersion: deleteMarkerVersion,
				})
			}
		}
	}

	s3util.WriteXML(w, r, http.StatusOK, marshallable)
}
