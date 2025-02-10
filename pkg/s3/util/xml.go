package s3util

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	s3error "github.com/jakthom/s3c/pkg/s3/error"
	zlog "github.com/rs/zerolog/log"
)

// WriteError serializes an error to a response as XML
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	s3Err := s3error.NewGenericError(r, err)
	WriteXML(w, r, s3Err.HTTPStatus, s3Err)
}

// writeXMLPrelude writes the HTTP headers and XML header to the response
func WriteXMLPrelude(w http.ResponseWriter, r *http.Request, code int) {
	vars := mux.Vars(r)
	requestID := vars["requestID"]

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amz-id-2", requestID)
	w.Header().Set("x-amz-request-id", requestID)
	w.WriteHeader(code)
	fmt.Fprint(w, xml.Header)
}

// writeXMLBody writes the marshaled XML payload of a value
func WriteXMLBody(w http.ResponseWriter, v interface{}) {
	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		// just log a message since a response has already been partially written
		zlog.Error().Err(err).Msg("failed to write XML response")
	}
}

// WriteXML writes HTTP headers, the XML header, and the XML payload to the
// response
func WriteXML(w http.ResponseWriter, r *http.Request, code int, v interface{}) {
	WriteXMLPrelude(w, r, code)
	WriteXMLBody(w, v)
}

// readXMLBody reads an HTTP request body's bytes, and unmarshals it into
// `payload`.
func ReadXMLBody(r *http.Request, payload interface{}) error {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(bodyBytes, &payload)
	if err != nil {
		return s3error.MalformedXMLError(r)
	}
	return nil
}
