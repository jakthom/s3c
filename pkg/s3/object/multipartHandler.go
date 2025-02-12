package s3object

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	s3error "github.com/jakthom/s3c/pkg/s3/error"
	s3user "github.com/jakthom/s3c/pkg/s3/user"
	s3util "github.com/jakthom/s3c/pkg/s3/util"
)

type MultipartHandler struct {
	Controller MultipartController
}

func (h *MultipartHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	keyMarker := r.FormValue("key-marker")
	uploadIDMarker := r.FormValue("upload-id-marker")
	if keyMarker == "" {
		uploadIDMarker = ""
	}

	maxUploads, err := s3util.IntFormValue(r, "max-uploads", 0, defaultMaxUploads, defaultMaxUploads)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	result, err := h.Controller.ListMultipart(r, bucket, keyMarker, uploadIDMarker, maxUploads)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	// some clients (e.g. minio-python) can't handle sub-seconds in datetime
	// output
	for _, upload := range result.Uploads {
		upload.Initiated = upload.Initiated.UTC().Round(time.Second)
	}

	marshallable := struct {
		XMLName            xml.Name  `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListMultipartUploadsResult"`
		Bucket             string    `xml:"Bucket"`
		KeyMarker          string    `xml:"KeyMarker"`
		UploadIDMarker     string    `xml:"UploadIdMarker"`
		NextKeyMarker      string    `xml:"NextKeyMarker"`
		NextUploadIDMarker string    `xml:"NextUploadIdMarker"`
		MaxUploads         int       `xml:"MaxUploads"`
		IsTruncated        bool      `xml:"IsTruncated"`
		Uploads            []*Upload `xml:"Upload"`
	}{
		Bucket:         bucket,
		KeyMarker:      keyMarker,
		UploadIDMarker: uploadIDMarker,
		MaxUploads:     maxUploads,
		IsTruncated:    result.IsTruncated,
		Uploads:        result.Uploads,
	}

	if marshallable.IsTruncated {
		highKey := ""
		highUploadID := ""

		for _, upload := range marshallable.Uploads {
			if upload.Key > highKey {
				highKey = upload.Key
			}
			if upload.UploadID > highUploadID {
				highUploadID = upload.UploadID
			}
		}

		marshallable.NextKeyMarker = highKey
		marshallable.NextUploadIDMarker = highUploadID
	}

	s3util.WriteXML(w, r, http.StatusOK, marshallable)
}

func (h *MultipartHandler) ListChunks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	maxParts, err := s3util.IntFormValue(r, "max-parts", 0, defaultMaxParts, defaultMaxParts)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	partNumberMarker, err := s3util.IntFormValue(r, "part-number-marker", 0, maxPartsAllowed, 0)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	uploadID := r.FormValue("uploadId")

	result, err := h.Controller.ListMultipartChunks(r, bucket, key, uploadID, partNumberMarker, maxParts)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	marshallable := struct {
		XMLName              xml.Name     `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListPartsResult"`
		Bucket               string       `xml:"Bucket"`
		Key                  string       `xml:"Key"`
		UploadID             string       `xml:"UploadId"`
		Initiator            *s3user.User `xml:"Initiator"`
		Owner                *s3user.User `xml:"Owner"`
		StorageClass         string       `xml:"StorageClass"`
		PartNumberMarker     int          `xml:"PartNumberMarker"`
		NextPartNumberMarker int          `xml:"NextPartNumberMarker"`
		MaxParts             int          `xml:"MaxParts"`
		IsTruncated          bool         `xml:"IsTruncated"`
		Parts                []*Part      `xml:"Part"`
	}{
		Bucket:           bucket,
		Key:              key,
		UploadID:         uploadID,
		PartNumberMarker: partNumberMarker,
		MaxParts:         maxParts,
		Initiator:        result.Initiator,
		Owner:            result.Owner,
		StorageClass:     result.StorageClass,
		IsTruncated:      result.IsTruncated,
		Parts:            result.Parts,
	}

	if marshallable.IsTruncated {
		high := 0

		for _, part := range marshallable.Parts {
			if part.PartNumber > high {
				high = part.PartNumber
			}
		}

		marshallable.NextPartNumberMarker = high
	}

	s3util.WriteXML(w, r, http.StatusOK, marshallable)
}

func (h *MultipartHandler) init(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID, err := h.Controller.InitMultipart(r, bucket, key)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	marshallable := struct {
		XMLName  xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ InitiateMultipartUploadResult"`
		Bucket   string   `xml:"Bucket"`
		Key      string   `xml:"Key"`
		UploadID string   `xml:"UploadId"`
	}{
		Bucket:   bucket,
		Key:      key,
		UploadID: uploadID,
	}

	s3util.WriteXML(w, r, http.StatusOK, marshallable)
}

func (h *MultipartHandler) Complete(w http.ResponseWriter, r *http.Request) {
	if err := s3util.RequireContentLength(r); err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID := r.FormValue("uploadId")

	payload := struct {
		XMLName xml.Name `xml:"CompleteMultipartUpload"`
		Parts   []*Part  `xml:"Part"`
	}{}
	if err := s3util.ReadXMLBody(r, &payload); err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	// verify that there's at least part, and all parts are in ascending order
	isSorted := sort.SliceIsSorted(payload.Parts, func(i, j int) bool {
		return payload.Parts[i].PartNumber < payload.Parts[j].PartNumber
	})
	if len(payload.Parts) == 0 || !isSorted {
		s3util.WriteError(w, r, s3error.InvalidPartOrderError(w, r))
		return
	}

	for _, part := range payload.Parts {
		part.ETag = s3util.AddETagQuotes(part.ETag)
	}

	ch := make(chan struct {
		result *CompleteMultipartResult
		err    error
	})

	go func() {
		result, err := h.Controller.CompleteMultipart(r, bucket, key, uploadID, payload.Parts)
		ch <- struct {
			result *CompleteMultipartResult
			err    error
		}{
			result: result,
			err:    err,
		}
	}()

	streaming := false

	for {
		select {
		case value := <-ch:
			if value.err != nil {
				s3Error := s3error.NewGenericError(r, value.err)

				if streaming {
					s3util.WriteXMLBody(w, s3Error)
				} else {
					s3util.WriteError(w, r, s3Error)
				}
			} else {
				marshallable := struct {
					XMLName  xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ CompleteMultipartUploadResult"`
					Location string   `xml:"Location"`
					Bucket   string   `xml:"Bucket"`
					Key      string   `xml:"Key"`
					ETag     string   `xml:"ETag"`
				}{
					Bucket:   bucket,
					Key:      key,
					Location: value.result.Location,
					ETag:     s3util.AddETagQuotes(value.result.ETag),
				}

				if value.result.Version != "" {
					w.Header().Set("x-amz-version-id", value.result.Version)
				}

				if streaming {
					s3util.WriteXMLBody(w, marshallable)
				} else {
					s3util.WriteXML(w, r, http.StatusOK, marshallable)
				}
			}
			return
		case <-time.After(completeMultipartPing):
			if !streaming {
				streaming = true
				s3util.WriteXMLPrelude(w, r, http.StatusOK)
			} else {
				fmt.Fprint(w, " ")
			}
		}
	}
}

func (h *MultipartHandler) Put(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID := r.FormValue("uploadId")
	partNumber, err := s3util.IntFormValue(r, "partNumber", 0, maxPartsAllowed, 0)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	etag, err := h.Controller.UploadMultipartChunk(r, bucket, key, uploadID, partNumber, r.Body)
	if err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	if etag != "" {
		w.Header().Set("ETag", s3util.AddETagQuotes(etag))
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MultipartHandler) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID := r.FormValue("uploadId")

	if err := h.Controller.AbortMultipart(r, bucket, key, uploadID); err != nil {
		s3util.WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
