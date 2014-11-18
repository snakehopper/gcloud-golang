// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"io"
	"time"

	raw "code.google.com/p/google-api-go-client/storage/v1"
	"golang.org/x/net/context"
)

// Owner represents the owner of a GCS object.
type Owner struct {
	// Entity identifies the owner, it's always in the form of "user-<userId>".
	Entity string `json:"entity,omitempty"`
}

// Bucket represents a Google Cloud Storage bucket.
type Bucket struct {
	// Name is the name of the bucket.
	Name string `json:"name,omitempty"`

	// ACL is the list of access control rules on the bucket.
	ACL []ACLRule `json:"acl,omitempty"`

	// DefaultObjectACL is the list of access controls to
	// apply to new objects when no object ACL is provided.
	DefaultObjectACL []ACLRule `json:"defaultObjectAcl,omitempty"`

	// Location is the location of the bucket. It defaults to "US".
	Location string `json:"location,omitempty"`

	// Metageneration is the metadata generation of the bucket.
	// Read-only.
	Metageneration int64 `json:"metageneration,omitempty"`

	// StorageClass is the storage class of the bucket. This defines
	// how objects in the bucket are stored and determines the SLA
	// and the cost of storage. Typical values are "STANDARD" and
	// "DURABLE_REDUCED_AVAILABILITY". Defaults to "STANDARD".
	StorageClass string `json:"storageClass,omitempty"`

	// Created is the creation time of the bucket.
	// Read-only.
	Created time.Time `json:"timeCreated,omitempty"`
}

func newBucket(b *raw.Bucket) *Bucket {
	if b == nil {
		return nil
	}
	bucket := &Bucket{
		Name:           b.Name,
		Location:       b.Location,
		Metageneration: b.Metageneration,
		StorageClass:   b.StorageClass,
		Created:        convertTime(b.TimeCreated),
	}
	acl := make([]ACLRule, len(b.Acl))
	for i, rule := range b.Acl {
		acl[i] = ACLRule{
			Entity: rule.Entity,
			Role:   ACLRole(rule.Role),
		}
	}
	bucket.ACL = acl
	objACL := make([]ACLRule, len(b.DefaultObjectAcl))
	for i, rule := range b.DefaultObjectAcl {
		objACL[i] = ACLRule{
			Entity: rule.Entity,
			Role:   ACLRole(rule.Role),
		}
	}
	bucket.DefaultObjectACL = objACL
	return bucket
}

// Object represents a Google Cloud Storage (GCS) object.
type Object struct {
	// Bucket is the name of the bucket containing this GCS object.
	Bucket string `json:"bucket,omitempty"`

	// Name is the name of the object.
	Name string `json:"name,omitempty"`

	// CacheControl control how long browser and Internet caches for the object.
	CacheControl string

	// ContentType is the MIME type of the object's content.
	ContentType string `json:"contentType,omitempty"`

	// ContentLanguage is the content language of the object's content.
	ContentLanguage string `json:"contentLanguage,omitempty"`

	// ACL is the list of access control rules for the object.
	ACL []ACLRule `json:"acl,omitempty"`

	// Owner is the owner of the object. Owner is alway the original
	// uploader of the object.
	// Read-only.
	Owner Owner `json:"owner,omitempty"`

	// Size is the length of the object's content.
	// Read-only.
	Size uint64 `json:"size,omitempty"`

	// ContentEncoding is the encoding of the object's content.
	// Read-only.
	ContentEncoding string `json:"contentEncoding,omitempty"`

	// MD5 is the MD5 hash of the data.
	// Read-only.
	MD5 []byte `json:"md5Hash,omitempty"`

	// CRC32C is the CRC32C checksum of the object's content.
	// Read-only.
	CRC32C []byte `json:"crc32c,omitempty"`

	// MediaLink is an URL to the object's content.
	// Read-only.
	MediaLink string `json:"mediaLink,omitempty"`

	// Metadata represents user-provided metadata, in key/value pairs.
	// It can be nil if no metadata is provided.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Generation is the generation version of the object's content.
	// Read-only.
	Generation int64 `json:"generation,omitempty"`

	// MetaGeneration is the version of the metadata for this
	// object at this generation. This field is used for preconditions
	// and for detecting changes in metadata. A metageneration number
	// is only meaningful in the context of a particular generation
	// of a particular object.
	// Read-only.
	MetaGeneration int64 `json:"metageneration,omitempty"`

	// StorageClass is the storage class of the object.
	// Read-only.
	StorageClass string `json:"storageClass,omitempty"`

	// Deleted is the deletion time of the object (or the zero-value time).
	// This will be non-zero if and only if this version of the object has been deleted.
	// Read-only.
	Deleted time.Time `json:"timeDeleted,omitempty"`

	// Updated is the creation or modification time of the object.
	// For buckets with versioning enabled, changing an object's
	// metadata does not change this property.
	// Read-only.
	Updated time.Time `json:"updated,omitempty"`
}

func (o *Object) toRawObject() *raw.Object {
	acl := make([]*raw.ObjectAccessControl, len(o.ACL))
	for i, rule := range o.ACL {
		acl[i] = &raw.ObjectAccessControl{
			Entity: rule.Entity,
			Role:   string(rule.Role),
		}
	}
	return &raw.Object{
		Bucket:          o.Bucket,
		Name:            o.Name,
		CacheControl:    o.CacheControl,
		ContentType:     o.ContentType,
		ContentEncoding: o.ContentEncoding,
		ContentLanguage: o.ContentLanguage,
		Acl:             acl,
		Metadata:        o.Metadata,
	}
}

// convertTime converts a time in RFC3339 format to time.Time.
// If any error occurs in parsing, the zero-value time.Time is silently returned.
func convertTime(t string) time.Time {
	var r time.Time
	if t != "" {
		r, _ = time.Parse(time.RFC3339, t)
	}
	return r
}

func newObject(o *raw.Object) *Object {
	if o == nil {
		return nil
	}
	acl := make([]ACLRule, len(o.Acl))
	for i, rule := range o.Acl {
		acl[i] = ACLRule{
			Entity: rule.Entity,
			Role:   ACLRole(rule.Role),
		}
	}
	return &Object{
		Bucket:          o.Bucket,
		Name:            o.Name,
		CacheControl:    o.CacheControl,
		ContentType:     o.ContentType,
		ContentLanguage: o.ContentLanguage,
		ACL:             acl,
		Owner:           Owner{Entity: o.Owner.Entity},
		ContentEncoding: o.ContentEncoding,
		Size:            o.Size,
		MD5:             []byte(o.Md5Hash),
		CRC32C:          []byte(o.Crc32c),
		MediaLink:       o.MediaLink,
		Metadata:        o.Metadata,
		Generation:      o.Generation,
		MetaGeneration:  o.Metageneration,
		StorageClass:    o.StorageClass,
		Deleted:         convertTime(o.TimeDeleted),
		Updated:         convertTime(o.Updated),
	}
}

// Query represents a query to filter objects from a bucket.
type Query struct {
	// Delimiter returns results in a directory-like fashion.
	// Results will contain only objects whose names, aside from the
	// prefix, do not contain delimiter. Objects whose names,
	// aside from the prefix, contain delimiter will have their name,
	// truncated after the delimiter, returned in prefixes.
	// Duplicate prefixes are omitted.
	// Optional.
	Delimiter string

	// Prefix is the prefix filter to query objects
	// whose names begin with this prefix.
	// Optional.
	Prefix string

	// Versions indicates whether multiple versions of the same
	// object will be included in the results.
	Versions bool

	// Cursor is a previously-returned page token
	// representing part of the larger set of results to view.
	// Optional.
	Cursor string

	// MaxResults is the maximum number of items plus prefixes
	// to return. As duplicate prefixes are omitted,
	// fewer total results may be returned than requested.
	// The default page limit is used if it is negative or zero.
	MaxResults int
}

// Objects represents a list of objects returned from
// a bucket look-p request and a query to retrieve more
// objects from the next pages.
type Objects struct {
	// Results represent a list of object results.
	Results []*Object

	// Next is the continuation query to retrieve more
	// results with the same filtering criteria. If there
	// are no more results to retrieve, it is nil.
	Next *Query

	// Prefixes represents prefixes of objects
	// matching-but-not-listed up to and including
	// the requested delimiter.
	Prefixes []string
}

// contentTyper implements ContentTyper to enable an
// io.ReadCloser to specify its MIME type.
type contentTyper struct {
	io.ReadCloser
	t string
}

func (c *contentTyper) ContentType() string {
	return c.t
}

// newObjectWriter returns a new ObjectWriter that writes to
// the file that is specified by info.Bucket and info.Name.
// Metadata changes are also reflected on the remote object
// entity, read-only fields are ignored during the write operation.
func newObjectWriter(ctx context.Context, info *Object) *ObjectWriter {
	w := &ObjectWriter{
		ctx:  ctx,
		done: make(chan bool),
	}
	pr, pw := io.Pipe()
	w.rc = &contentTyper{pr, info.ContentType}
	w.pw = pw
	go func() {
		resp, err := rawService(ctx).Objects.Insert(
			info.Bucket, info.toRawObject()).Media(w.rc).Do()
		w.err = err
		if err == nil {
			w.obj = newObject(resp)
		}
		close(w.done)
	}()
	return w
}

// ObjectWriter is an io.WriteCloser that opens a connection
// to update the metadata and file contents of a GCS object.
type ObjectWriter struct {
	ctx context.Context

	rc io.ReadCloser
	pw *io.PipeWriter

	done chan bool
	obj  *Object
	err  error
}

// Write writes len(p) bytes to the object. It returns the number
// of the bytes written, or an error if there is a problem occured
// during the write. It's a blocking operation, and will not return
// until the bytes are written to the underlying socket.
func (w *ObjectWriter) Write(p []byte) (n int, err error) {
	if w.err != nil {
		return 0, w.err
	}
	return w.pw.Write(p)
}

// Close closes the writer and cleans up other resources
// used by the writer.
func (w *ObjectWriter) Close() error {
	if w.err != nil {
		return w.err
	}
	w.rc.Close()
	return w.pw.Close()
}

// Object returns the object information. It will block until
// the write operation is complete.
func (w *ObjectWriter) Object() (*Object, error) {
	<-w.done
	return w.obj, w.err
}
