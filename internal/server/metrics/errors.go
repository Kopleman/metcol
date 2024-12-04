package metrics

import "errors"

var ErrCounterValueParse = errors.New("can not parse counter value from store")
var ErrGougeValueParse = errors.New("can not parse gouge value from store")
var ErrValueParse = errors.New("can not parse input value")
var ErrUnknownMetricType = errors.New("unknown metric type")
var ErrStoreKeyParse = errors.New("unable to parse store key")
