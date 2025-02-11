package s3bucket

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	s3object "github.com/jakthom/s3c/pkg/s3/object"
	s3util "github.com/jakthom/s3c/pkg/s3/util"
)

type BucketHandler struct {
	Controller BucketController
}

func (h *BucketHandler) Location(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	location, err := h.Controller.GetLocation(r, bucket)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	s3util.WriteXML(w, r, http.StatusOK, struct {
		XMLName  xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ LocationConstraint"`
		Location string   `xml:",innerxml"`
	}{
		Location: location,
	})
}

func (h *BucketHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	// hadoop's S3 client could make requests with max-keys=5000
	maxKeys, err := s3util.IntFormValue(r, "max-keys", 0, 5000, DefaultMaxKeys)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	prefix := r.FormValue("prefix")
	marker := r.FormValue("marker")
	delimiter := r.FormValue("delimiter")

	result, err := h.Controller.ListObjects(r, bucket, prefix, marker, delimiter, maxKeys)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	// some clients (e.g. minio-python) can't handle sub-seconds in datetime
	// output
	for _, contents := range result.Contents {
		contents.LastModified = contents.LastModified.UTC().Round(time.Second)
	}

	for _, c := range result.Contents {
		c.ETag = s3util.AddETagQuotes(c.ETag)
	}

	marshallable := struct {
		XMLName        xml.Name                   `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListBucketResult"`
		Contents       []*s3object.Object         `xml:"Contents"` // TODO -> should it actually be named contents? IDK
		CommonPrefixes []*s3object.CommonPrefixes `xml:"CommonPrefixes"`
		Delimiter      string                     `xml:"Delimiter,omitempty"`
		IsTruncated    bool                       `xml:"IsTruncated"`
		Marker         string                     `xml:"Marker"`
		MaxKeys        int                        `xml:"MaxKeys"`
		Name           string                     `xml:"Name"`
		NextMarker     string                     `xml:"NextMarker,omitempty"`
		Prefix         string                     `xml:"Prefix"`
	}{
		Name:           bucket,
		Prefix:         prefix,
		Marker:         marker,
		Delimiter:      delimiter,
		MaxKeys:        maxKeys,
		IsTruncated:    result.IsTruncated,
		Contents:       result.Contents,
		CommonPrefixes: result.CommonPrefixes,
	}

	if marshallable.IsTruncated {
		high := ""

		for _, contents := range marshallable.Contents {
			if contents.Key > high {
				high = contents.Key
			}
		}
		for _, commonPrefix := range marshallable.CommonPrefixes {
			if commonPrefix.Prefix > high {
				high = commonPrefix.Prefix
			}
		}

		marshallable.NextMarker = high
	}

	s3util.WriteXML(w, r, http.StatusOK, marshallable)
}

func (h *BucketHandler) Put(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if err := h.Controller.CreateBucket(r, bucket); err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BucketHandler) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if err := h.Controller.DeleteBucket(r, bucket); err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// func (h *BucketHandler) versioning(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	bucket := vars["bucket"]

// 	status, err := h.controller.GetBucketVersioning(r, bucket)
// 	if err != nil {
// 		s3util.WriteError(w, r, err)
// 		return
// 	}

// 	result := struct {
// 		XMLName xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ VersioningConfiguration"`
// 		Status  string   `xml:"Status,omitempty"`
// 	}{
// 		Status: status,
// 	}

// 	s3util.WriteXML(w, r, http.StatusOK, result)
// }

// func (h *BucketHandler) setVersioning(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	bucket := vars["bucket"]

// 	payload := struct {
// 		XMLName xml.Name `xml:"VersioningConfiguration"`
// 		Status  string   `xml:"Status"`
// 	}{}
// 	if err := s3util.ReadXMLBody(r, &payload); err != nil {
// 		s3util.WriteError(w, r, err)
// 		return
// 	}

// 	if payload.Status != VersioningDisabled && payload.Status != VersioningSuspended && payload.Status != VersioningEnabled {
// 		s3util.WriteError(w, r, s3error.IllegalVersioningConfigurationError(r))
// 		return
// 	}

// 	err := h.controller.SetBucketVersioning(r, bucket, payload.Status)
// 	if err != nil {
// 		s3util.WriteError(w, r, err)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// }

// func (h *BucketHandler) listVersions(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	bucket := vars["bucket"]

// 	maxKeys, err := s3util.IntFormValue(r, "max-keys", 0, DefaultMaxKeys, DefaultMaxKeys)
// 	if err != nil {
// 		s3util.WriteError(w, r, err)
// 		return
// 	}

// 	prefix := r.FormValue("prefix")
// 	keyMarker := r.FormValue("key-marker")
// 	versionIDMarker := r.FormValue("version-id-marker")
// 	delimiter := r.FormValue("delimiter")

// 	result, err := h.controller.ListObjectVersions(r, bucket, prefix, keyMarker, versionIDMarker, delimiter, maxKeys)
// 	if err != nil {
// 		s3util.WriteError(w, r, err)
// 		return
// 	}

// 	// some clients (e.g. minio-python) can't handle sub-seconds in datetime
// 	// output
// 	for _, version := range result.Versions {
// 		version.LastModified = version.LastModified.UTC().Round(time.Second)
// 	}
// 	for _, deleteMarker := range result.DeleteMarkers {
// 		deleteMarker.LastModified = deleteMarker.LastModified.UTC().Round(time.Second)
// 	}

// 	for _, v := range result.Versions {
// 		v.ETag = s3util.AddETagQuotes(v.ETag)
// 	}

// 	marshallable := struct {
// 		XMLName             xml.Name                 `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListVersionsResult"`
// 		Delimiter           string                   `xml:"Delimiter,omitempty"`
// 		IsTruncated         bool                     `xml:"IsTruncated"`
// 		KeyMarker           string                   `xml:"KeyMarker"`
// 		NextKeyMarker       string                   `xml:"NextKeyMarker,omitempty"`
// 		MaxKeys             int                      `xml:"MaxKeys"`
// 		Name                string                   `xml:"Name"`
// 		VersionIDMarker     string                   `xml:"VersionIdKeyMarker"`
// 		NextVersionIDMarker string                   `xml:"NextVersionIdKeyMarker,omitempty"`
// 		Prefix              string                   `xml:"Prefix"`
// 		Versions            []*Version               `xml:"Version"`
// 		DeleteMarkers       []*s3object.DeleteMarker `xml:"DeleteMarker"`
// 	}{
// 		IsTruncated:     result.IsTruncated,
// 		KeyMarker:       keyMarker,
// 		MaxKeys:         maxKeys,
// 		Name:            bucket,
// 		VersionIDMarker: versionIDMarker,
// 		Prefix:          prefix,
// 		Versions:        result.Versions,
// 		DeleteMarkers:   result.DeleteMarkers,
// 	}

// 	if marshallable.IsTruncated {
// 		highKey := ""
// 		highVersion := ""

// 		for _, version := range marshallable.Versions {
// 			if version.Key > highKey {
// 				highKey = version.Key
// 			}
// 			if version.Version > highVersion {
// 				highVersion = version.Version
// 			}
// 		}
// 		for _, deleteMarker := range marshallable.DeleteMarkers {
// 			if deleteMarker.Key > highKey {
// 				highKey = deleteMarker.Key
// 			}
// 			if deleteMarker.Version > highVersion {
// 				highVersion = deleteMarker.Version
// 			}
// 		}

// 		marshallable.NextKeyMarker = highKey
// 		marshallable.NextVersionIDMarker = highVersion
// 	}

// 	s3util.WriteXML(w, r, http.StatusOK, marshallable)
// }
