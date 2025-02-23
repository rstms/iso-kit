package info

// ImageObject is an interface that represents an object in an ISO9660 image.
// It is used to provide generic information about objects from the file system image which allows tools like 'isoview'
// to display information about the image layout.
type ImageObject interface {
	Type() string
	Name() string
	Description() string
	Properties() map[string]interface{}
	Offset() int64
	Size() int
	GetObjects() []ImageObject
	Marshal() ([]byte, error)
}
